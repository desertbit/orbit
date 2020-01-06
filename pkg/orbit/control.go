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
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/packet"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/rs/zerolog"
)

const (
	// A default error message that is sent back to a caller of remote function,
	// in case the Func of the remote did not return an error that conforms to
	// our Error interface.
	// This is done to prevent sensitive information to leak to the outside that
	// is usually carried in normal errors.
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

type control struct {
	s *Session

	closingChan <-chan struct{}
	codec       codec.Codec
	log         *zerolog.Logger

	streamWriteMx sync.Mutex
	stream        net.Conn

	callRetChain *chain
}

func newControl(s *Session, stream net.Conn, log *zerolog.Logger) *control {
	return &control{
		s:            s,
		closingChan:  s.ClosingChan(),
		codec:        s.cf.Codec,
		log:          log,
		stream:       stream,
		callRetChain: newChain(),
	}
}

func (c *control) call()

func (c *control) write(reqType byte, headerI interface{}, dataI interface{}) (err error) {
	var header, payload []byte

	// Marshal the header data.
	header, err = msgpack.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("encode header: %v", err)
	}

	// Marshal the payload data with the configured codec,
	// unless the payload is a byte slice. Use that directly.
	if dataI != nil {
		switch v := dataI.(type) {
		case []byte:
			payload = v

		default:
			payload, err = c.codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("encode: %v", err)
			}
		}
	}

	// Ensure only one write happens at a time.
	c.streamWriteMx.Lock()
	defer c.streamWriteMx.Unlock()

	// Set the deadline when all write operations must be finished.
	// todo: timeout
	/*err = c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	if err != nil {
		return err
	}*/

	// Write the request type.
	_, err = c.stream.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	// todo: timeout
	err = packet.Write(c.stream, header, 0)
	if err != nil {
		return err
	}

	// Write the payload.
	// todo: timeout
	err = packet.Write(c.stream, payload, 0)
	if err != nil {
		return err
	}

	return nil
}

// cancelCall sends a request to the remote peer in order to cancel an ongoing request
// identified by the given key.
func (c *control) cancelCall(key uint64) {
	err := c.write(
		typeCallCancel,
		&api.ControlCancel{Key: key},
		nil,
	)
	if err != nil {
		c.log.Error().Err(err).Msg("control cancel call")
	}
}

// readRoutine listens on the control stream and reads the packets from it.
// It expects each request to consist of one byte (the request type), followed
// by a packet (see https://github.com/desertbit/orbit/pkg/packet), thus, a header
// and optionally a payload.
//
// This function blocks in an endless loop, it should therefore run in its
// own goroutine.
//
// It recovers from panics and logs errors with the configured logger of
// the Control.
//
// If a request could be successfully read, it is passed to the
// handleRequest() method to process it.
func (c *control) readRoutine(conn net.Conn, once bool) {
	// Close the control on exit.
	defer conn.Close()

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
			err = fmt.Errorf("catched panic: %v", e)
		}

		// Only log if not closed.
		if err != nil && err != io.EOF && !c.s.IsClosing() {
			c.log.Printf("control: read: %v", err)
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
			n, err = c.stream.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		// Read the header from the stream.
		headerData, err = packet.Read(c.stream, nil, 0)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(c.stream, nil, 0)
		if err != nil {
			return
		}

		// Handle the received message in a new goroutine.
		if once {
			c.handleRequest(reqType, headerData, payloadData)
			return
		} else {
			go c.handleRequest(reqType, headerData, payloadData)
		}
	}
}

// waitForResponse waits for a response on the given chainChan channel.
// Returns ErrClosed, if the control is closed before the response arrives.
// Returns ErrCallTimeout, if the timeout exceeds before the response arrives.
// Returns ErrCallCanceled, if the call is canceled by the caller.
func (c *control) waitForResponse(
	ctx context.Context,
	key uint64,
	channel chainChan,
) (octx *Data, err error) {
	// Wait for a response.
	select {
	case <-c.closingChan:
		// Abort, if the control closes.
		err = ErrClosed

	case <-ctx.Done():
		c.cancelCall(key)

	case rData := <-channel:
		// Response has arrived.
		octx = rData.Context
		err = rData.Err
	}
	return
}

// handleRequest handles an incoming request read by the readRoutine() function
// and decides how to proceed with it.
// It does this based on the request type byte, which indicates right now only
// whether the data must be processed as Call or CallReturn.
//
// Panics are recovered and wrapped in an error.
// Any error is logged using the control logger.
func (c *control) handleRequest(reqType byte, headerData, payloadData []byte) {
	var err error

	// Catch panics, caused by the handler func or one of the hooks.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
		if err != nil {

			c.log.Printf("control: handleRequest: %v", err)
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = c.handleCall(headerData, payloadData)
	case typeCallReturn:
		err = c.handleCallReturn(headerData, payloadData)
	case typeCallCancel:
		err = c.handleCallCancel(headerData)
	default:
		err = fmt.Errorf("invalid request type: %v", reqType)
	}
}

// handleCall processes an incoming request with request type 'typeCall'.
// It decodes the header and uses it to retrieve the correct handler
// function for this request.
// The handler function is then executed with the payload.
// If no error occurs, the response is sent back to the caller.
//
// If defined, the call hook is called during the execution.
// If defined, the error hook is called at the end, in case
// an error occurred.
//
// This function is executed on the side of the receiver of a request,
// the "server-side".
func (c *control) handleCall(headerData, payloadData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCall
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode header call: %v", err)
	}

	// Retrieve the handler function for this request.
	c.s.callFuncsMx.RLock()
	f, ok := c.s.callFuncs[header.ID]
	c.s.callFuncsMx.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exist: id=%v", header.ID)
	}

	// Build the request data for the handler function.
	ctx := newData(payloadData, c.codec)

	// Save the context in our active contexts map, if the request is cancelable.
	// todo: cancel
	/*if header.Cancelable {
		c.activeCallContextsMx.Lock()
		c.activeCallContexts[header.Key] = ctx
		c.activeCallContextsMx.Unlock()

		// Ensure to remove the context from the map.
		defer func() {
			c.activeCallContextsMx.Lock()
			delete(c.activeCallContexts, header.Key)
			c.activeCallContextsMx.Unlock()
		}()
	}*/

	// Call the call hook, if defined.
	// todo: hook
	/*if c.callHook != nil {
		c.callHook(c, header.ID, ctx)
	}*/

	var (
		msg  string
		code int
	)

	// Execute the handler function.
	retData, retErr := f(ctx)
	if retErr != nil {
		// Decide what to send back to the caller.
		if cErr, ok := retErr.(Error); ok {
			code = cErr.Code()
			msg = cErr.Msg()
		} else if c.s.cf.SendErrToCaller {
			msg = retErr.Error()
		}

		// Ensure an error message is always set.
		if msg == "" {
			msg = defaultErrorMessage
		}

		// Log the actual error here, but only if it contains a message.
		if retErr.Error() != "" {
			c.log.Error().Err(retErr).Str("callID", header.ID).Msg("handle call")
		}

		// Call the error hook, if defined.
		// todo: hook
		/*if c.errorHook != nil {
			c.errorHook(c, header.ID, retErr)
		}*/
	}

	// Skip the return if this is a one way call.
	if header.Key == 0 {
		return nil
	}

	// Build the header for the response.
	retHeader := &api.ControlReturn{
		Key:  header.Key,
		Msg:  msg,
		Code: code,
	}

	// Send the response back to the caller.
	err = c.write(typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("call request: send response: %v", err)
	}

	return nil
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
func (c *control) handleCallReturn(headerData, payloadData []byte) (err error) {
	// Decode the header.
	var header api.ControlReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode header call return: %v", err)
	}

	// Get the channel by the key.
	channel := c.callRetChain.get(header.Key)
	if channel == nil {
		return fmt.Errorf("return request failed: no return channel set (call timeout exceeded?)")
	}

	// Create a new context.
	ctx := newData(payloadData, c.codec)

	// Create the channel data.
	rData := chainData{Context: ctx}

	// Create an ErrorCode, if an error is present.
	if header.Msg != "" {
		rData.Err = &ErrorCode{err: header.Msg, Code: header.Code}
	}

	// Send the return data to the channel.
	// Ensure that there is a receiving endpoint.
	// Otherwise we would have a lost blocking goroutine.
	select {
	case channel <- rData:
		return nil

	default:
		// Retry with a timeout.
		timeout := time.NewTimer(time.Second)
		defer timeout.Stop()

		select {
		case channel <- rData:
			return nil
		case <-timeout.C:
			return fmt.Errorf("return request failed (call timeout exceeded?)")
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
func (c *control) handleCallCancel(headerData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCancel
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode header call cancel: %v", err)
	}

	// Retrieve the context from the active contexts map and delete
	// it right away from the map again to ensure that a context is
	// canceled exactly once.
	var (
		//ctx *Data
		ok bool
	)
	/*
		// todo: cancel
		c.activeCallContextsMx.Lock()
		ctx, ok = c.activeCallContexts[header.Key]
		delete(c.activeCallContexts, header.Key)
		c.activeCallContextsMx.Unlock()*/

	// If there is no context available for this key, do nothing.
	if !ok {
		return
	}

	// Cancel the currently running request.
	// Since we deleted the context from the map already, this code
	// is executed exactly once and does not block on the buffered channel.
	// todo: cancel
	//ctx.cancelChan <- struct{}{}
	return
}
