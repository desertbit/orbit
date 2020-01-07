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

package orbit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/packet"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
)

const (
	// A default error message that is sent back to a caller of remote function,
	// in case the Func of the remote did not return an error that conforms to
	// our Error interface.
	// This is done to prevent sensitive information to leak to the outside that
	// is usually carried in normal errors.
	// TODO: Rename to defaultCallErrorMessage
	defaultErrorMessage = "method call failed"

	// The first byte send in a request to indicate a call to a remote function.
	typeCall byte = 0

	// The first byte send in a request to indicate the response from a remote
	// called function.
	typeCallReturn byte = 1

	// The first byte send in a request to indicate that a currently running
	// request should be canceled.
	typeCallCancel byte = 2
)

func (s *Session) RegisterCall(id string, f CallFunc) {
	s.callFuncsMx.Lock()
	s.callFuncs[id] = f
	s.callFuncsMx.Unlock()
}

func (s *Session) Call(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	return s.call(ctx, s.callStream, id, data)
}

func (s *Session) CallAsync(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	stream, err := s.openStream(ctx, "", api.StreamTypeCallAsync)
	if err != nil {
		return
	}

	return s.call(ctx, newMxStream(stream), id, data)
}

func (s *Session) HandleCallAsync(stream net.Conn) {
	go s.readCallRoutine(newMxStream(stream), true)
}

//###############//
//### Private ###//
//###############//

func (s *Session) call(ctx context.Context, ms *mxStream, id string, data interface{}) (d *Data, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := s.callRetChain.new()
	defer s.callRetChain.delete(key)

	// Write to the client.
	err = s.writeCall(ctx, ms, typeCall, &api.ControlCall{ID: id, Key: key}, data)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return s.waitForCallResponse(ctx, ms, key, channel)
}

func (s *Session) writeCall(ctx context.Context, ms *mxStream, reqType byte, headerI, dataI interface{}) (err error) {
	var header, payload []byte

	// Marshal the header data.
	header, err = msgpack.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("control write encode header: %v", err)
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
				return fmt.Errorf("control write encode payload: %v", err)
			}
		}
	}

	// Ensure only one write happens at a time on the stream.
	ms.WriteMx.Lock()
	defer ms.WriteMx.Unlock()

	// Set the deadline when all write operations must be finished.
	deadline, ok := ctx.Deadline()
	if ok {
		err = ms.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	}

	// Write the request type.
	_, err = ms.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	err = packet.Write(ms, header)
	if err != nil {
		return err
	}

	// Write the payload.
	err = packet.Write(ms, payload)
	if err != nil {
		return err
	}

	// Reset the deadline.
	return ms.SetWriteDeadline(time.Time{})
}

// waitForCallResponse waits for a response on the given chainChan channel.
func (s *Session) waitForCallResponse(
	ctx context.Context,
	ms *mxStream,
	key uint32,
	channel chainChan,
) (d *Data, err error) {
	// Wait for a response.
	select {
	case <-s.ClosingChan():
		// Abort, if the control closes.
		err = ErrClosed
	case <-ctx.Done():
		// Cancel the call on the remote peer.
		err = s.writeCall(ctx, ms, typeCallCancel, &api.ControlCancel{Key: key}, nil)
	case rData := <-channel:
		// Response has arrived.
		d = rData.Data
		err = rData.Err
	}
	return
}

func (s *Session) readCallRoutine(ms *mxStream, once bool) {
	// Close the control on exit.
	defer ms.Close()

	// Warning: don't shadow the error.
	// Otherwise the deferred logging won't work!
	var (
		err          error
		bytesRead, n int
		reqTypeBuf   = make([]byte, 1)
	)

	// Catch panics and log error messages. There should be exactly one
	// readRoutine running all the time, therefore this defer does not
	// hurt performance at all and is a safety net, to prevent the server
	// from crashing, should anything panic during the reads.
	defer func() {
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: %v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: %v", e)
			}
		}

		// Only log if not closed.
		if err != nil && !errors.Is(err, io.EOF) && !s.IsClosing() {
			s.log.Error().
				Err(err).
				Msg("control read")
		}
	}()

	for {
		// This variables must be redefined for each loop,
		// because this data is used in the new goroutine.
		// Otherwise there is a race and data gets corrupted.
		var (
			reqType                 byte
			headerData, payloadData []byte
		)

		// Read the reqType from the stream.
		// Read in a loop, as Read could potentially return 0 read bytes.
		bytesRead = 0
		for bytesRead == 0 {
			n, err = ms.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		// Read the header from the stream.
		headerData, err = packet.Read(ms, nil)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(ms, nil)
		if err != nil {
			return
		}

		// Handle the received message in a new goroutine.
		if once {
			s.handleCallRequest(ms, reqType, headerData, payloadData)
			return
		} else {
			go s.handleCallRequest(ms, reqType, headerData, payloadData)
		}
	}
}

func (s *Session) handleCallRequest(ms *mxStream, reqType byte, headerData, payloadData []byte) {
	var err error

	// Catch panics, caused by the handler func or one of the hooks.
	defer func() {
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: %v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: %v", e)
			}
		}

		if err != nil {
			s.log.Error().
				Err(err).
				Msg("control handle request")
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = s.handleCall(ms, headerData, payloadData)
	case typeCallReturn:
		err = s.handleCallReturn(headerData, payloadData)
	case typeCallCancel:
		err = s.handleCallCancel(headerData)
	default:
		err = fmt.Errorf("invalid request type '%v'", reqType)
	}
}

func (s *Session) handleCall(ms *mxStream, headerData, payloadData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCall
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call decode header: %v", err)
	}

	// Retrieve the handler function for this request.
	s.callFuncsMx.RLock()
	f, ok := s.callFuncs[header.ID]
	s.callFuncsMx.RUnlock()
	if !ok {
		return fmt.Errorf("handle call: func '%v' does not exist", header.ID)
	}

	// Build the request data for the handler function.
	data := newData(payloadData, s.codec)

	// Save a context in our active contexts map so we can cancel it, if needed.
	// TODO: context is not meant for canceling, see https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// TODO: lets enhance our closer and bring it to the next level
	cc := newCallContext()
	s.callActiveCtxsMx.Lock()
	s.callActiveCtxs[header.Key] = cc
	s.callActiveCtxsMx.Unlock()

	// Ensure to remove the context from the map and to cancel it.
	defer func() {
		s.callActiveCtxsMx.Lock()
		delete(s.callActiveCtxs, header.Key)
		s.callActiveCtxsMx.Unlock()
		cc.cancel()
	}()

	// Call the call hook, if defined.
	// TODO: hook
	/*if s.callHook != nil {
		s.callHook(c, header.ID, ctx)
	}*/

	var (
		msg  string
		code int
	)

	// Execute the handler function.
	retData, retErr := f(cc.ctx, s, data)
	if retErr != nil {
		// Decide what to send back to the caller.
		var cErr Error
		if errors.As(err, &cErr) {
			code = cErr.Code()
			msg = cErr.Msg()
		} else if s.cf.SendErrToCaller {
			msg = retErr.Error()
		}

		// Ensure an error message is always set.
		if msg == "" {
			msg = defaultErrorMessage
		}

		// Log the actual error here, but only if it contains a message.
		if retErr.Error() != "" {
			s.log.Error().
				Err(retErr).
				Str("func", header.ID).
				Msg("control handle call")
		}

		// Call the error hook, if defined.
		// TODO: hook
		/*if s.errorHook != nil {
			s.errorHook(c, header.ID, retErr)
		}*/
	}

	// Build the header for the response.
	retHeader := &api.ControlReturn{
		Key:  header.Key,
		Msg:  msg,
		Code: code,
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeCallReturnTimeout)
	defer cancel()

	// Send the response back to the caller.
	err = s.writeCall(ctx, ms, typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("handle call: send response: %v", err)
	}

	return
}

// handleCallReturn processes an incoming response with request type 'typeCallReturn'.
// It decodes the header and uses it to retrieve the correct channel for this callReturn.
// The response payload is then wrapped in a context and sent over the channel
// back to the calling function that waits for it.
// It can then deliver the response to the caller, which is the end of
// one request-response cycle.
//
// This function is executed on the side of the receiver of a response,
// the "client-side".
func (s *Session) handleCallReturn(headerData, payloadData []byte) (err error) {
	// Decode the header.
	var header api.ControlReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call return decode header: %v", err)
	}

	// Get the channel by the key.
	channel := s.callRetChain.get(header.Key)
	if channel == nil {
		return errors.New("call return: no handler func available")
	}

	// Create the channel data.
	rData := chainData{Data: newData(payloadData, s.codec)}

	// Create an ErrorCode, if an error is present.
	if header.Msg != "" {
		rData.Err = &ErrorCode{err: header.Msg, Code: header.Code}
	}

	// Send the return data to the channel.
	// Ensure that there is a receiving endpoint.
	// Otherwise we would have a lost blocking goroutine.
	select {
	case channel <- rData:
		return

	default:
		// Retry with a timeout.
		timeout := time.NewTimer(time.Second)
		defer timeout.Stop()

		select {
		case channel <- rData:
			return
		case <-timeout.C:
			return fmt.Errorf("call return: failed to deliver return data (timeout)")
		}
	}
}

// handleCallCancel processes an incoming request with request type 'typeCallCancel'.
// It decodes the header to retrieve the key of the request that should be canceled.
// It then retrieves the context of said request and closes its associated closer.
// If no request with the sent key could be found, nothing happens.
//
// This function is executed on the side of the receiver of a request,
// the "server-side".
func (s *Session) handleCallCancel(headerData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCancel
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call cancel decode header: %v", err)
	}

	// Retrieve the context from the active contexts map and delete
	// it right away from the map again to ensure that a context is
	// canceled exactly once.
	var (
		cc *callContext
		ok bool
	)

	// TODO: context is not meant for canceling, see https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// TODO: lets enhance our closer and bring it to the next level
	s.callActiveCtxsMx.Lock()
	cc, ok = s.callActiveCtxs[header.Key]
	delete(s.callActiveCtxs, header.Key)
	s.callActiveCtxsMx.Unlock()

	// If there is no context available for this key, do nothing.
	if !ok {
		return
	}

	// Cancel the currently running request.
	// Since we deleted the context from the map already, this code
	// is executed exactly once.
	cc.cancel()
	return
}
