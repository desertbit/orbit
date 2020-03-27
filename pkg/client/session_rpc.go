/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/rpc"
	"github.com/desertbit/orbit/pkg/transport"
)

func (s *session) Call(ctx context.Context, id string, arg, ret interface{}) (err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := s.chain.New()
	defer s.chain.Delete(key)

	// Create a new client context.
	cctx := newContext(ctx, s)

	// Call the OnCall hooks.
	err = s.handler.hookOnCall(cctx, id, key)
	if err != nil {
		return
	}

	ctxDone := ctx.Done()

	// Call the OnCallDone or OnCallCanceled hooks.
	defer func() {
		select {
		case <-ctxDone:
			s.handler.hookOnCallCanceled(cctx, id, key)
		default:
			s.handler.hookOnCallDone(cctx, id, key, err)
		}
	}()

	// Write to the client.
	err = s.writeRPCRequest(ctx, s.stream, &s.streamWriteMx, api.RPCTypeCall, &api.RPCCall{
		ID:   id,
		Key:  key,
		Data: cctx.header,
	}, arg)
	if err != nil {
		return
	}

	// Wait for the response and return its result.
	select {
	case <-s.ClosingChan():
		return ErrClosed

	case <-ctxDone:
		// Cancel the call on the remote peer.
		err = s.cancelCall(key)
		if err != nil {
			s.log.Error().
				Err(err).
				Uint32("key", key).
				Msg("rpc: call: failed to cancel call")
		}
		return ctx.Err()

	case r := <-channel:
		// Response has arrived. Check the error first.
		if r.Err != nil {
			return r.Err
		}

		// Skip the decoding of the return data if no data to decode to is passed.
		if ret == nil {
			return
		}

		// Data must be present if ret is set.
		if len(r.Data) == 0 {
			return ErrNoData
		}

		// Decode.
		return s.codec.Decode(r.Data, ret)
	}
}

func (s *session) AsyncCall(ctx context.Context, id string, arg, ret interface{}) (err error) {
	// Open a new stream for this call.
	stream, err := s.openStream(ctx, "", api.StreamTypeAsyncCall)
	if err != nil {
		return
	}
	defer stream.Close()

	// Create a new unique key. This will be used for cancelation.
	key := s.chain.NewKey()

	// Create a new client context.
	cctx := newContext(ctx, s)

	// Call the OnCall hooks.
	err = s.handler.hookOnCall(cctx, id, key)
	if err != nil {
		return
	}

	ctxDone := ctx.Done()

	// Call the OnCallDone or OnCallCanceled hooks.
	defer func() {
		select {
		case <-ctxDone:
			s.handler.hookOnCallCanceled(cctx, id, key)
		default:
			s.handler.hookOnCallDone(cctx, id, key, err)
		}
	}()

	var (
		errChan   = make(chan error, 1)
		dataChan  = make(chan []byte, 1)
		closeChan = make(chan struct{})
	)
	defer close(closeChan)

	// Write and receive the reponse in a new goroutine.
	// This enables to cancel the request as early as possible.
	go func() {
		// Ensure a new error variable is used.
		var err error

		// Write to the client. A locker is not required.
		err = s.writeRPCRequest(ctx, stream, nil, api.RPCTypeCall, &api.RPCCall{
			ID:   id,
			Key:  key,
			Data: cctx.header,
		}, arg)
		if err != nil {
			errChan <- err
			return
		}

		// Early return if the body function closed already
		// or the context was canceled.
		select {
		case <-closeChan:
			return
		case <-ctxDone:
			return
		default:
		}

		// Set a read deadline to the stream if present.
		if deadline, ok := ctx.Deadline(); ok {
			err = stream.SetReadDeadline(deadline)
			if err != nil {
				errChan <- err
				return
			}
		}

		// Wait for the response.
		reqType, headerData, payloadData, err := rpc.Read(stream, nil, nil)
		if err != nil {
			errChan <- err
			return
		} else if reqType != api.RPCTypeReturn {
			errChan <- fmt.Errorf("async call: invalid request type '%v'", reqType)
			return
		}

		// Early return if the body function closed already
		// or the context was canceled.
		select {
		case <-closeChan:
			return
		case <-ctxDone:
			return
		default:
		}

		// Decode the header.
		var header api.RPCReturn
		err = api.Codec.Decode(headerData, &header)
		if err != nil {
			errChan <- fmt.Errorf("async return request: decode header: %w", err)
			return
		}

		// Return an ErrorCode, if an error is present.
		if header.Err != "" {
			err = errImpl{msg: header.Err, code: header.ErrCode}
			errChan <- err
			return
		}

		dataChan <- payloadData
	}()

	// Wait for the data or a cancel event.
	select {
	case <-s.ClosingChan():
		return ErrClosed

	case <-ctxDone:
		// Cancel the call on the remote peer.
		err = s.cancelCall(key)
		if err != nil {
			s.log.Error().
				Err(err).
				Uint32("key", key).
				Msg("rpc: async call: failed to cancel call")
		}
		return ctx.Err()

	case err = <-errChan:
		return err

	case data := <-dataChan:
		// Skip the decoding of the return data if no data to decode to is passed.
		if ret == nil {
			return nil
		}

		// Data must be present if ret is set.
		if len(data) == 0 {
			return ErrNoData
		}

		// Decode.
		return s.codec.Decode(data, ret)
	}
}

func (s *session) writeRPCRequest(
	ctx context.Context,
	stream transport.Stream,
	streamLocker sync.Locker,
	reqType api.RPCType,
	headerI, dataI interface{},
) (err error) {
	var header, payload []byte

	// Check if already canceled.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Marshal the header data.
	header, err = api.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("write request: encode header: %w", err)
	}

	// Marshal the payload data with the configured codec,
	// unless the payload is a byte slice. Use that directly.
	if dataI != nil {
		switch v := dataI.(type) {
		case []byte:
			payload = v

		default:
			payload, err = s.codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("write request: encode payload: %w", err)
			}
		}
	}

	// Ensure only one write happens at a time on the stream.
	if streamLocker != nil {
		streamLocker.Lock()
		defer streamLocker.Unlock()
	}

	// Set the deadline when all write operations must be finished.
	deadline, ok := ctx.Deadline()
	if ok {
		err = stream.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	} else {
		err = stream.SetWriteDeadline(time.Time{})
		if err != nil {
			return err
		}
	}

	// Write the rpc request.
	return rpc.Write(stream, reqType, header, payload)
}
