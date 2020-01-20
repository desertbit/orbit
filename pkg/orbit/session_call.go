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
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/rs/zerolog/log"
)

const (
	// A default error message that is sent back to a caller of remote function,
	// in case the Func of the remote did not return an error that conforms to
	// our Error interface.
	// This is done to prevent sensitive information to leak to the outside that
	// is usually carried in normal errors.
	defaultCallErrorMessage = "method call failed"

	// The first byte send in a request to indicate a call to a remote function.
	typeCall byte = 0

	// The first byte send in a request to indicate the response from a remote
	// called function.
	typeCallReturn byte = 1

	// The first byte send in a request to indicate that a currently running
	// request should be canceled.
	typeCallCancel byte = 2
)

func (s *Session) RegisterCall(service, id string, f CallFunc) {
	s.callFuncsMx.Lock()
	s.callFuncs[service+"."+id] = f
	s.callFuncsMx.Unlock()
}

func (s *Session) Call(ctx context.Context, service, id string, data interface{}) (d *Data, err error) {
	// Retrieve the single call stream per service for this basic call.
	var (
		cs *callStream
		ok bool
	)

	s.callStreamsMx.Lock()
	cs, ok = s.callStreams[service]
	s.callStreamsMx.Unlock()

	if !ok {
		// First call, initialize the call stream.
		var stream net.Conn
		stream, err = s.openStream(ctx, "", api.StreamTypeCallInit)
		if err != nil {
			return
		}

		// Create new call stream.
		cs = newCallStream(stream, newChain())

		// Save the call stream for the provided service id.
		s.callStreamsMx.Lock()
		s.callStreams[service] = cs
		s.callStreamsMx.Unlock()

		// Start a read routine to receive responses and revcalls.
		go s.readCallRoutine(cs, false)
	}

	return s.call(ctx, cs, service+"."+id, data)
}

func (s *Session) CallAsync(ctx context.Context, service, id string, data interface{}) (d *Data, err error) {
	// Open a new stream connection.
	stream, err := s.openStream(ctx, "", api.StreamTypeCallAsync)
	if err != nil {
		return
	}

	// Create a temporary call stream, which lives only until the return data arrives.
	cs := newCallStream(stream, newChain())

	// Start a read routine to receive the response.
	go s.readCallRoutine(cs, true)

	return s.call(ctx, cs, service+"."+id, data)
}

//###############//
//### Private ###//
//###############//

func (s *Session) call(ctx context.Context, cs *callStream, id string, data interface{}) (d *Data, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := cs.RetChain.new()
	defer cs.RetChain.delete(key)

	// Write to the client.
	err = s.writeCall(ctx, cs, typeCall, &api.Call{ID: id, Key: key}, data)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return s.waitForCallResponse(ctx, cs, key, channel)
}

func (s *Session) writeCall(ctx context.Context, cs *callStream, reqType byte, headerI, dataI interface{}) (err error) {
	var header, payload []byte

	// Marshal the header data.
	header, err = msgpack.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("call: write encode header: %v", err)
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
				return fmt.Errorf("call: write encode payload: %v", err)
			}
		}
	}

	// Ensure only one write happens at a time on the stream.
	cs.WriteMx.Lock()
	defer cs.WriteMx.Unlock()

	// Set the deadline when all write operations must be finished.
	deadline, ok := ctx.Deadline()
	if ok {
		err = cs.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	}

	// Write the request type.
	_, err = cs.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	err = packet.Write(cs, header)
	if err != nil {
		return err
	}

	// Write the payload.
	err = packet.Write(cs, payload)
	if err != nil {
		return err
	}

	// Reset the deadline.
	return cs.SetWriteDeadline(time.Time{})
}

// waitForCallResponse waits for a response on the given chainChan channel.
func (s *Session) waitForCallResponse(
	ctx context.Context,
	cs *callStream,
	key uint32,
	channel chainChan,
) (d *Data, err error) {
	// Wait for a response.
	select {
	case <-s.ClosingChan():
		// Abort, if the session closes.
		err = ErrClosed
	case <-ctx.Done():
		err = ctx.Err()
		if errors.Is(err, context.Canceled) {
			// Ignore the canceled error, the caller knows it himself.
			err = nil
		}

		// Cancel the call on the remote peer.
		wErr := s.writeCall(ctx, cs, typeCallCancel, &api.CallCancel{Key: key}, nil)
		if wErr != nil {
			if err == nil {
				err = wErr
			} else {
				log.Error().
					Err(wErr).
					Uint32("key", key).
					Msg("wait for call response: failed to write call cancel")
			}
		}
	case rData := <-channel:
		// Response has arrived.
		d = rData.Data
		err = rData.Err
	}
	return
}

func (s *Session) handleAsyncCall(stream net.Conn) {
	// No need to create a chain for return values. Not used on the server side.
	go s.readCallRoutine(newCallStream(stream, nil), true)
}

func (s *Session) handleInitCall(stream net.Conn) {
	// A fixed call stream, create a chain for it.
	go s.readCallRoutine(newCallStream(stream, newChain()), false)
}

func (s *Session) readCallRoutine(cs *callStream, once bool) {
	// Close the stream on exit.
	defer cs.Close()

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
				Msg("call read error")
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
			n, err = cs.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		// Read the header from the stream.
		headerData, err = packet.Read(cs, nil)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(cs, nil)
		if err != nil {
			return
		}

		if once {
			// Handle the received message in the same routine.
			s.handleCallRequest(cs, reqType, headerData, payloadData)
			return
		} else {
			// Handle the received message in a new goroutine.
			go s.handleCallRequest(cs, reqType, headerData, payloadData)
		}
	}
}

func (s *Session) handleCallRequest(cs *callStream, reqType byte, headerData, payloadData []byte) {
	var err error

	// Catch panics, caused by the handler func or one of the hooks.
	defer func() {
		// TODO: Remove panic handling?, execCallHandler handles panics on its own and other code is only ours, must not panic.
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: \n%v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: \n%v", e)
			}
		}

		if err != nil {
			// Log. Do not use the Err() field, as stack trace formatting is lost then.
			s.log.Error().
				Msgf("session: failed to handle call request: \n%v", err)
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = s.handleCall(cs, headerData, payloadData)
	case typeCallReturn:
		err = s.handleCallReturn(cs, headerData, payloadData)
	case typeCallCancel:
		err = s.handleCallCancel(headerData)
	default:
		err = fmt.Errorf("invalid request type '%v'", reqType)
	}
}

func (s *Session) handleCall(cs *callStream, headerData, payloadData []byte) (err error) {
	// Decode the request header.
	var header api.Call
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call: decode header: %v", err)
	}

	// Execute the handler for this call.
	retHeader, retData := s.execCallHandler(header.Key, header.ID, payloadData)

	ctx, cancel := context.WithTimeout(context.Background(), writeCallReturnTimeout)
	defer cancel()

	// Send the response back to the caller.
	err = s.writeCall(ctx, cs, typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("handle call: send response: %v", err)
	}

	return
}

// Sets the msg and code in the retHeader, in case of an error.
// Recovers from panics.
func (s *Session) execCallHandler(key uint32, id string, payloadData []byte) (retHeader *api.CallReturn, retData interface{}) {
	// Build the header for the response.
	retHeader = &api.CallReturn{Key: key}

	var err error
	defer func() {
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: \n%v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: \n%v", e)
			}
		}

		if err != nil {
			// Log. Do not use the Err() field, as stack trace formatting is lost then.
			s.log.Error().
				Str("func", id).
				Msgf("exec call handler: \n%v", err)

			// Ensure an error message is always set.
			if retHeader.Msg == "" {
				if s.cf.SendErrToCaller {
					retHeader.Msg = err.Error()
				} else {
					retHeader.Msg = defaultCallErrorMessage
				}
			}
		}
	}()

	// Retrieve the handler function for this request.
	var (
		ok bool
		f  CallFunc
	)
	s.callFuncsMx.RLock()
	f, ok = s.callFuncs[id]
	s.callFuncsMx.RUnlock()
	if !ok {
		err = errors.New("call handler not found")
		return
	}

	// Save a context in our active contexts map so we can cancel it, if needed.
	// TODO: context is not meant for canceling, see https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// TODO: lets enhance our closer and bring it to the next level
	cc := newCallContext()
	s.callActiveCtxsMx.Lock()
	s.callActiveCtxs[retHeader.Key] = cc
	s.callActiveCtxsMx.Unlock()

	// Ensure to remove the context from the map and to cancel it.
	defer func() {
		s.callActiveCtxsMx.Lock()
		delete(s.callActiveCtxs, retHeader.Key)
		s.callActiveCtxsMx.Unlock()
		cc.cancel()
	}()

	// Execute the handler function.
	retData, err = f(cc.ctx, s, newData(payloadData, s.codec))
	if err != nil {
		// Check, if an orbit error was returned.
		var cErr Error
		if errors.As(err, &cErr) {
			retHeader.Code = cErr.Code()
			retHeader.Msg = cErr.Msg()
		}
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
func (s *Session) handleCallReturn(cs *callStream, headerData, payloadData []byte) (err error) {
	// Decode the header.
	var header api.CallReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call return decode header: %v", err)
	}

	// Get the channel by the key.
	channel := cs.RetChain.get(header.Key)
	if channel == nil {
		return fmt.Errorf("call return: no handler func available for key '%d'", header.Key)
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
	var header api.CallCancel
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
