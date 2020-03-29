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

package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/rpc"
	"github.com/desertbit/orbit/pkg/transport"
)

func (s *session) writeRPCRequest(
	ctx context.Context,
	stream transport.Stream,
	streamLocker sync.Locker,
	reqType api.RPCType,
	headerI, dataI interface{},
	maxPayloadSize int,
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
		return fmt.Errorf("encode header: %w", err)
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
				return fmt.Errorf("encode payload: %w", err)
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
	return rpc.Write(stream, reqType, header, payload, s.maxHeaderSize, maxPayloadSize)
}

func (s *session) startRPCReadRoutine() {
	go s.rpcReadRoutine()
}

func (s *session) rpcReadRoutine() {
	// Close the session on exit.
	defer s.Close_()

	for {
		reqType, header, payload, err := rpc.Read(s.stream, nil, nil, s.maxHeaderSize, s.maxArgSize)
		if err != nil {
			// Log errors, but only, if the session or stream are not closing.
			if !s.IsClosing() && !s.stream.IsClosed() && !s.conn.IsClosedError(err) {
				s.log.Error().
					Err(err).
					Msg("rpc: read routine")
			}
			return
		}

		// Handle the request in a new routine.
		go s.handleRPCRequest(reqType, header, payload)
	}
}

func (s *session) handleRPCRequest(reqType api.RPCType, header, payload []byte) {
	var err error

	// Check the request type.
	// The service supports a different range of requests than the client.
	switch reqType {
	case api.RPCTypeCall:
		err = s.handleCall(s.stream, &s.streamWriteMx, header, payload, s.maxRetSize)
	default:
		err = fmt.Errorf("invalid request type '%v'", reqType)
	}
	if err != nil {
		s.log.Error().
			Err(err).
			Msg("rpc: failed to handle request")
	}
}

func (s *session) handleCall(
	stream transport.Stream,
	streamLocker sync.Locker,
	header []byte,
	payload []byte,
	maxRetSize int,
) (err error) {
	// Decode the request header.
	var h api.RPCCall
	err = api.Codec.Decode(header, &h)
	if err != nil {
		return fmt.Errorf("call: decode header: %w", err)
	}

	// Get the call.
	c, err := s.handler.getCall(h.ID)
	if err != nil {
		return fmt.Errorf("call %s: %w", h.ID, err)
	}

	// Prepare our return header.
	retHeader := api.RPCReturn{
		Key: h.Key,
	}

	// Create a context for cancelation and add the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Call the hooks and function in a nested function.
	// We must pass the error from the function call to the done hook.
	ret, err := func() (ret interface{}, err error) {
		// Create the service context.
		sctx := newContext(ctx, s, h.Data)

		// Call the OnCall hooks.
		err = s.handler.hookOnCall(sctx, h.ID, h.Key)
		if err != nil {
			return nil, fmt.Errorf("call %s: %w", h.ID, err)
		}

		// Call the OnCallDone or OnCallCanceled hooks.
		defer func() {
			select {
			case <-ctx.Done():
				s.handler.hookOnCallCanceled(sctx, h.ID, h.Key)
			default:
				s.handler.hookOnCallDone(sctx, h.ID, h.Key, err)
			}
		}()

		// Publish the cancel function, so the client can cancel it with the key.
		alreadyCanceled := s.setCancelFunc(h.Key, cancel)
		if alreadyCanceled {
			cancel()
			// The call has been already canceled.
			// Cancel requests may be handled earlier than the actual call request.
			return nil, fmt.Errorf("call %s: canceled early", h.ID)
		}

		// Always remove the cancel function.
		defer s.deleteCancelFunc(h.Key)

		// Call the actual call handler.
		return s.handler.handleCall(sctx, c.f, payload)
	}()
	if err != nil {
		// Check, if an orbit error was returned.
		var oErr Error
		if errors.As(err, &oErr) {
			retHeader.ErrCode = oErr.Code()
			retHeader.Err = oErr.Msg()
		}

		// Ensure an error message is always set.
		if retHeader.Err == "" {
			if s.sendInternalErrors {
				retHeader.Err = err.Error()
			} else {
				retHeader.Err = fmt.Sprintf("%s call failed", h.ID)
			}
		}

		// Reset the error, because we handled it already and the result should be send to the caller.
		err = nil
	}

	// Send the response back to the caller.
	err = s.writeRPCRequest(ctx, stream, streamLocker, api.RPCTypeReturn, retHeader, ret, maxRetSize)
	if err != nil {
		return fmt.Errorf("call %s: write response: %w", h.ID, err)
	}
	return
}
