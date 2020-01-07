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
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
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
	closer.Closer

	s            *Session
	log          *zerolog.Logger
	codec        codec.Codec
	main         *controlStream
	callRetChain *chain

	activeCtxsMx sync.Mutex
	activeCtxs   map[uint32]*controlContext
}

func newControl(s *Session, stream net.Conn) *control {
	c := &control{
		Closer:       s,
		s:            s,
		codec:        s.cf.Codec,
		log:          s.log,
		main:         newControlStream(stream),
		callRetChain: newChain(),
		activeCtxs:   make(map[uint32]*controlContext),
	}

	// Start the read routine.
	go c.readRoutine(c.main, false)

	return c
}

func (c *control) Call(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	return c.call(ctx, c.main, id, data)
}

func (c *control) CallAsync(ctx context.Context, stream net.Conn, id string, data interface{}) (d *Data, err error) {
	return c.call(ctx, newControlStream(stream), id, data)
}

func (c *control) HandleCallAsync(stream net.Conn) {
	go c.readRoutine(newControlStream(stream), true)
}

//###############//
//### Private ###//
//###############//

func (c *control) call(ctx context.Context, cs *controlStream, id string, data interface{}) (d *Data, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := c.callRetChain.new()
	defer c.callRetChain.delete(key)

	// Write to the client.
	err = c.write(ctx, cs, typeCall, &api.ControlCall{ID: id, Key: key}, data)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return c.waitForResponse(ctx, cs, key, channel)
}

func (c *control) write(ctx context.Context, cs *controlStream, reqType byte, headerI, dataI interface{}) (err error) {
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
			payload, err = c.codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("control write encode payload: %v", err)
			}
		}
	}

	// Ensure only one write happens at a time on the stream.
	cs.writeMx.Lock()
	defer cs.writeMx.Unlock()

	// Set the deadline when all write operations must be finished.
	deadline, ok := ctx.Deadline()
	if ok {
		err = cs.stream.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	}

	// Write the request type.
	_, err = cs.stream.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	err = packet.Write(cs.stream, header)
	if err != nil {
		return err
	}

	// Write the payload.
	err = packet.Write(cs.stream, payload)
	if err != nil {
		return err
	}

	// Reset the deadline.
	return cs.stream.SetWriteDeadline(time.Time{})
}

// waitForResponse waits for a response on the given chainChan channel.
func (c *control) waitForResponse(
	ctx context.Context,
	cs *controlStream,
	key uint32,
	channel chainChan,
) (d *Data, err error) {
	// Wait for a response.
	select {
	case <-c.ClosingChan():
		// Abort, if the control closes.
		err = ErrClosed
	case <-ctx.Done():
		// Cancel the call on the remote peer.
		err = c.write(ctx, cs, typeCallCancel, &api.ControlCancel{Key: key}, nil)
	case rData := <-channel:
		// Response has arrived.
		d = rData.Data
		err = rData.Err
	}
	return
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
func (c *control) readRoutine(cs *controlStream, once bool) {
	// Close the control on exit.
	defer cs.stream.Close()

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
		// TODO: print stack trace...
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}

		// Only log if not closed.
		if err != nil && !errors.Is(err, io.EOF) && !c.s.IsClosing() {
			c.log.Error().
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
			n, err = cs.stream.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		// Read the header from the stream.
		headerData, err = packet.Read(cs.stream, nil)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(cs.stream, nil)
		if err != nil {
			return
		}

		// Handle the received message in a new goroutine.
		if once {
			c.handleRequest(cs, reqType, headerData, payloadData)
			return
		} else {
			go c.handleRequest(cs, reqType, headerData, payloadData)
		}
	}
}

// handleRequest handles an incoming request read by the readRoutine() function
// and decides how to proceed with it.
// It does this based on the request type byte, which indicates right now only
// whether the data must be processed as Call or CallReturn.
//
// Panics are recovered and wrapped in an error.
// Any error is logged using the control logger.
func (c *control) handleRequest(cs *controlStream, reqType byte, headerData, payloadData []byte) {
	var err error

	// Catch panics, caused by the handler func or one of the hooks.
	defer func() {
		// TODO: print stack trace...
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
		if err != nil {
			c.log.Error().
				Err(err).
				Msg("control")
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = c.handleCall(cs, headerData, payloadData)
	case typeCallReturn:
		err = c.handleCallReturn(headerData, payloadData)
	case typeCallCancel:
		err = c.handleCallCancel(headerData)
	default:
		err = fmt.Errorf("handle request: invalid request type '%v'", reqType)
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
func (c *control) handleCall(cs *controlStream, headerData, payloadData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCall
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call decode header: %v", err)
	}

	// Retrieve the handler function for this request.
	c.s.callFuncsMx.RLock()
	f, ok := c.s.callFuncs[header.ID]
	c.s.callFuncsMx.RUnlock()
	if !ok {
		return fmt.Errorf("handle call: func '%v' does not exist", header.ID)
	}

	// Build the request data for the handler function.
	data := newData(payloadData, c.codec)

	// Save a context in our active contexts map so we can cancel it, if needed.
	// TODO: context is not meant for canceling, see https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// TODO: lets enhance our closer and bring it to the next level
	cc := newControlContext()
	c.activeCtxsMx.Lock()
	c.activeCtxs[header.Key] = cc
	c.activeCtxsMx.Unlock()

	// Ensure to remove the context from the map and to cancel it.
	defer func() {
		c.activeCtxsMx.Lock()
		delete(c.activeCtxs, header.Key)
		c.activeCtxsMx.Unlock()
		cc.cancel()
	}()

	// Call the call hook, if defined.
	// TODO: hook
	/*if c.callHook != nil {
		c.callHook(c, header.ID, ctx)
	}*/

	var (
		msg  string
		code int
	)

	// Execute the handler function.
	retData, retErr := f(cc.ctx, c.s, data)
	if retErr != nil {
		// Decide what to send back to the caller.
		var cErr Error
		if errors.As(err, &cErr) {
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
			c.log.Error().
				Err(retErr).
				Str("func", header.ID).
				Msg("control handle call")
		}

		// Call the error hook, if defined.
		// TODO: hook
		/*if c.errorHook != nil {
			c.errorHook(c, header.ID, retErr)
		}*/
	}

	// Build the header for the response.
	retHeader := &api.ControlReturn{
		Key:  header.Key,
		Msg:  msg,
		Code: code,
	}

	// TODO: don;t pass the context to write! If canceled, then tell this the client!

	// Send the response back to the caller.
	// TODO: set an own server timeout here?
	err = c.write(cc.ctx, cs, typeCallReturn, retHeader, retData)
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
func (c *control) handleCallReturn(headerData, payloadData []byte) (err error) {
	// Decode the header.
	var header api.ControlReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("handle call return decode header: %v", err)
	}

	// Get the channel by the key.
	channel := c.callRetChain.get(header.Key)
	if channel == nil {
		return errors.New("call return: no handler func available")
	}

	// Create the channel data.
	rData := chainData{Data: newData(payloadData, c.codec)}

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
func (c *control) handleCallCancel(headerData []byte) (err error) {
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
		cc *controlContext
		ok bool
	)

	// TODO: context is not meant for canceling, see https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// TODO: lets enhance our closer and bring it to the next level
	c.activeCtxsMx.Lock()
	cc, ok = c.activeCtxs[header.Key]
	delete(c.activeCtxs, header.Key)
	c.activeCtxsMx.Unlock()

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
