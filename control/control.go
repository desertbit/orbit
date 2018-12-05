/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
Package control provides an implementation of a simple network protocol that
offers a RPC-like request/response mechanism between two peers.

Each peer can register functions on his control that the remote peer can
then call, either synchronously with a response, or asynchronously.
Making calls is possible by using the Call- methods defined in this package.

A Control must be created once. Then it is advised to register all
possible requests on it. If the configuration is done, the Ready()
method must be called on it to start the routine that accepts incoming requests.

Network Data Format

Control uses the packet package (https://github.com/desertbit/orbit/packet)
to send and receive data over the connection.

Encoding

Each packet's header is automatically encoded with the msgpack.Codec
(https://github.com/desertbit/orbit/codec/msgpack). The payloads
are encoded with the codec.Codec defined in its configuration file.
In case the payload is already a slice of bytes, the encoding is skipped.

Hooks

The Control offers a call- and an errorHook. These are useful for
logging purposes and similar tasks. Check out their documentation
for further information on when exactly they are called.

Error Handling

On the 'server' side, each handler function registered to
handle a certain request may return an error that satisfies the
control.Error interface, to specifically determine, what the client
receives as message and to specify error codes the client can react to.

On the 'client' side, the Call- functions return (beside the usual errors
that can happen) the ErrorCode struct that satisfies the standard
error interface, but contains in addition the error code the server can
set to indicate certain errors. Clients should therefore always check,
whether the returned error is of type ErrorCode, to handle certain errors
appropriately.
*/
package control

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/packet"
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
)

var (
	// ErrClosed defines the error if the socket connection is closed.
	ErrClosed = errors.New("socket closed")

	// ErrCallTimeout defines the error if the call timeout is reached.
	ErrCallTimeout = errors.New("call timeout")

	// ErrWriteTimeout defines the error if a write operation's timeout is reached.
	ErrWriteTimeout = errors.New("write timeout")
)

// The Func type defines a callable Control function.
// If the returned error is not nil and implements the Error
// interface of this package, the clients can react to
// predefined error codes and act accordingly.
type Func func(ctx *Context) (data interface{}, err error)

// The Funcs type defines a set of callable Control functions.
type Funcs map[string]Func

// The CallHook type defines the callback function that is executed for every
// Func called by the remote peer.
type CallHook func(c *Control, funcID string, ctx *Context)

// The ErrorHook type defines the error callback function that is executed
// for every Func calley by the remote peer, that produces an error.
type ErrorHook func(c *Control, funcID string, err error)

// The Control type is the main struct of this package
// that is used to implement RPC.
type Control struct {
	closer.Closer

	// Stores the configuration for the Control, such as timeouts, codecs, etc.
	config *Config
	// The logger used to log messages.
	logger *log.Logger

	// Synchronises write operations to the network connection.
	connWriteMutex sync.Mutex
	// The raw network connection that is used to send the requests over.
	conn net.Conn

	// Contains a channel for each request that is used to deliver the
	// response back to the correct calling function.
	callRetChain *chain
	// Synchronises the access to the handler function map.
	funcMapMutex sync.RWMutex
	// Stores the handler functions for the requests. The key to the map
	// is the id of the request.
	funcMap map[string]Func

	// Called for every incoming request. Can be useful for logging
	// purposes or similar tasks.
	callHook CallHook
	// Called for every incoming request, if during the execution
	// an error occurs. Can be useful for logging purposes or similar tasks.
	errorHook ErrorHook
}

// New creates a new Control using the passed connection.
//
// A config can be passed to manage the behaviour of the
// Control. Any value of the config that has not been set,
// a default value is provided. If a nil config is passed,
// a config with default values is created.
//
// Ready() must be called on the Control to start its read routine,
// so it can accept incoming requests.
func New(conn net.Conn, config *Config) *Control {
	// Create a new socket.
	config = prepareConfig(config)
	s := &Control{
		Closer:       closer.New(),
		config:       config,
		logger:       config.Logger,
		conn:         conn,
		callRetChain: newChain(),
		funcMap:      make(map[string]Func),
	}
	s.OnClose(conn.Close)

	return s
}

// Logger returns the log.Logger of the control's config.
func (c *Control) Logger() *log.Logger {
	return c.logger
}

// Codec returns the codec.Codec of the control's config.
func (c *Control) Codec() codec.Codec {
	return c.config.Codec
}

// Ready signalizes the Control that the initialization is done.
// The socket starts reading from the underlying connection.
// This should be called only once per Control.
func (c *Control) Ready() {
	// Start the service routines.
	go c.readRoutine()
}

// LocalAddr returns the local network address of the Control.
func (c *Control) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address of the peer
// the Control is connected to.
func (c *Control) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetCallHook sets the call hook function which is triggered, if a local
// remote callable function will be called. This hook can be used for logging purpose.
// Only set this hook during initialization (before Ready() has been called).
func (c *Control) SetCallHook(h CallHook) {
	c.callHook = h
}

// SetErrorHook sets the error hook function which is triggered, if a local
// remote callable function returns an error. This hook can be used for logging purpose.
// Only set this hook during initialization (before Ready() has been called).
func (c *Control) SetErrorHook(h ErrorHook) {
	c.errorHook = h
}

// AddFunc registers a local function that can be called by the remote peer.
// This method is thread-safe and may be called at any time.
func (c *Control) AddFunc(id string, f Func) {
	c.funcMapMutex.Lock()
	c.funcMap[id] = f
	c.funcMapMutex.Unlock()
}

// AddFuncs registers a map of local functions that can be called by the remote peer.
// This method is thread-safe and may be called at any time.
func (c *Control) AddFuncs(funcs Funcs) {
	c.funcMapMutex.Lock()
	// Iterate through the map and register the functions.
	for id, f := range funcs {
		c.funcMap[id] = f
	}
	c.funcMapMutex.Unlock()
}

// Call is a convenience function that calls CallTimeout(), but uses the default
// CallTimeout from the config of the Control.
func (c *Control) Call(id string, data interface{}) (*Context, error) {
	return c.CallTimeout(id, data, c.config.CallTimeout)
}

// CallTimeout calls a remote function and waits for its result. The given id determines,
// which function should be called on the remote. The given timeout determines how long
// the request/response process may take at a maximum.
//
// The passed data are sent as payload of the request and get automatically encoded
// with the codec.Codec of the Control.
// If data is nil, no payload is sent.
// If data is a byte slice, the encoding is skipped.
//
// This method blocks until the remote function sent back a response and returns the
// Context of this response. It is therefore considered 'synchronous'.
//
// Returns ErrCallTimeout on a timeout.
// Returns ErrClosed if the connection is closed.
//
// This method is thread-safe.
func (c *Control) CallTimeout(
	id string,
	data interface{},
	timeout time.Duration,
) (ctx *Context, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := c.callRetChain.new()
	defer c.callRetChain.delete(key)

	// Create the header.
	header := &api.ControlCall{
		ID:  id,
		Key: key,
	}

	// Write to the client.
	err = c.write(typeCall, header, data)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return c.waitForResponse(timeout, channel)
}

// CallOneWay calls a remote function, but the remote peer will not send
// back a response and this func will immediately return, as soon as
// the data has been written to the connection.
//
// This method is thread-safe.
func (c *Control) CallOneWay(id string, data interface{}) error {
	return c.CallAsync(id, data, nil)
}

// CallAsync calls a remote function in an asynchronous fashion, as the
// response will be awaited in a new goroutine and passed to the given callback.
// It uses the default CallTimeout from the config of the Control.
//
// This method is thread-safe.
func (c *Control) CallAsync(
	id string,
	data interface{},
	callback func(ctx *Context, err error),
) error {
	return c.CallAsyncTimeout(id, data, c.config.CallTimeout, callback)
}

// CallAsync calls a remote function in an asynchronous fashion, as the
// response will be awaited in a new goroutine and passed to the given callback.
//
// The response will be awaited in a new goroutine. The given callback will
// receive an ErrCallTimeout error, should the timeout be exceeded.
//
// This method is thread-safe.
func (c *Control) CallAsyncTimeout(
	id string,
	data interface{},
	timeout time.Duration,
	callback func(ctx *Context, err error),
) error {
	var (
		key     uint64
		channel chainChan
	)

	// If a callback has been defined, create a channel that will be used to send
	// the data over that forms the response to the call.
	if callback != nil {
		key, channel = c.callRetChain.new()
	}

	// Create the header.
	header := &api.ControlCall{
		ID:  id,
		Key: key,
	}

	// Send the request.
	err := c.write(typeCall, header, data)
	if err != nil {
		return err
	}

	// No response awaited, quit since it is a one way call.
	if callback == nil {
		return nil
	}

	// Wait for the response, but in an new routine.
	go callback(c.waitForResponse(timeout, channel))

	return nil
}

//###############//
//### Private ###//
//###############//

// write sends a packet to the remote peer that
// consists of the given header and payload
// data. The packet is preceded by one byte, which
// indicates the type of request we are sending
// (call or callReturn, see constants).
//
// Per specification of the packet format
// (see https://github.com/desertbit/orbit/packet), the
// header must always be present, while the payload is optional.
//
// Both header and payload are encoded, with the exception that,
// if the payload is not nil and of type []byte, the encoding
// of it is skipped.
//
// Returns ErrWriteTimeout, if the deadline of the write is not met.
// Returns packet.ErrMaxPayloadSizeExceeded, if either the header
// or the payload exceed the MaxMessageSize defined in the config.
func (c *Control) write(reqType byte, headerI interface{}, dataI interface{}) (err error) {
	var header, payload []byte

	// Marshal the header data.
	header, err = msgpack.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("encode header: %v", err)
	}

	// Marshal the payload data if present
	// or use the direct byte slice if set.
	if dataI != nil {
		switch v := dataI.(type) {
		case []byte:
			payload = v

		default:
			payload, err = c.config.Codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("encode: %v", err)
			}
		}
	}

	// Check the size of the header and the payload beforehand, so that we do not
	// write something onto the connection and then fail.
	if len(header) > c.config.MaxMessageSize || len(payload) > c.config.MaxMessageSize {
		return packet.ErrMaxPayloadSizeExceeded
	}

	// Ensure only one write happens at a time.
	c.connWriteMutex.Lock()
	defer c.connWriteMutex.Unlock()

	// Set the deadline when all write operations must be finished.
	err = c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	if err != nil {
		return err
	}

	// In case the write timeouts, assign to it our ErrWriteTimeout variable.
	defer func() {
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				err = ErrWriteTimeout
			}
		}
	}()

	// Write the request type.
	_, err = c.conn.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	err = packet.Write(c.conn, header, c.config.MaxMessageSize)
	if err != nil {
		return err
	}

	// Write the payload.
	err = packet.Write(c.conn, payload, c.config.MaxMessageSize)
	if err != nil {
		return err
	}

	return nil
}

// waitForResponse waits for a response on the given chainChan channel.
// Returns ErrClosed, if the control is closed before the response arrives.
// Returns ErrCallTimeout, if the given timeout is exceeded
// before the response arrives.
func (c *Control) waitForResponse(
	timeout time.Duration,
	channel chainChan,
) (ctx *Context, err error) {
	// Create the timeout.
	timeoutTimer := time.NewTimer(timeout)

	// Wait for a response.
	select {
	case <-c.CloseChan():
		// Abort if the Control closes.
		err = ErrClosed

	case <-timeoutTimer.C:
		// Abort if the deadline is over.
		err = ErrCallTimeout

	case rData := <-channel:
		// Response has arrived.
		ctx = rData.Context
		err = rData.Err
	}

	// Stop the timeout.
	_ = timeoutTimer.Stop()

	return
}

// readRoutine listens on the control stream and reads the packets from it.
// It expects each request to consist of one byte (the request type), followed
// by a packet (see https://github.com/desertbit/orbit/packet), thus, a header
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
func (c *Control) readRoutine() {
	// Close the control on exit.
	defer c.Close()

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
		if err != nil && err != io.EOF && !c.IsClosed() {
			c.logger.Printf("control: read: %v", err)
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

		// No timeout, as we need to wait here for any incoming request.
		err = c.conn.SetReadDeadline(time.Time{})
		if err != nil {
			return
		}

		// Read the reqType from the stream.
		// Read in a loop, as Read could potentially return 0 read bytes.
		bytesRead = 0
		for bytesRead == 0 {
			n, err = c.conn.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		// Set a single read deadline for both read operations.
		err = c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
		if err != nil {
			return
		}

		// Read the header from the stream.
		headerData, err = packet.Read(c.conn, nil, c.config.MaxMessageSize)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(c.conn, nil, c.config.MaxMessageSize)
		if err != nil {
			return
		}

		// Handle the received message in a new goroutine.
		go func() {
			gerr := c.handleRequest(reqType, headerData, payloadData)
			if gerr != nil {
				c.logger.Printf("control: handleRequest: %v", gerr)
			}
		}()
	}
}

// handleRequest handles an incoming request read by the readRoutine() function
// and decides how to proceed with it.
// It does this based on the request type byte, which indicates right now only
// whether the data must be processed as Call or CallReturn.
//
// Panics are recovered and wrapped in an error.
// If the request type is unknown, an error is returned.
func (c *Control) handleRequest(reqType byte, headerData, payloadData []byte) (err error) {
	// Catch panics, caused by the handler func or one of the hooks.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = c.handleCall(headerData, payloadData)

	case typeCallReturn:
		err = c.handleCallReturn(headerData, payloadData)

	default:
		err = fmt.Errorf("invalid request type: %v", reqType)
	}
	return
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
func (c *Control) handleCall(headerData, payloadData []byte) (err error) {
	// Decode the request header.
	var header api.ControlCall
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode control call: %v", err)
	}

	// Retrieve the handler function for this request.
	c.funcMapMutex.RLock()
	f, ok := c.funcMap[header.ID]
	c.funcMapMutex.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exist: id=%v", header.ID)
	}

	// Build the request context for the handler function.
	ctx := newContext(c, payloadData)

	// Call the call hook, if defined.
	if c.callHook != nil {
		c.callHook(c, header.ID, ctx)
	}

	var (
		code int
		msg  string
	)

	// Execute the handler function.
	retData, retErr := f(ctx)
	if retErr != nil {
		// Decide what to send back to the caller.
		if cErr, ok := retErr.(Error); ok {
			code = cErr.Code()
			msg = cErr.Msg()
		} else if c.config.SendErrToCaller {
			msg = retErr.Error()
		}

		// Ensure an error message is always set.
		if msg == "" {
			msg = defaultErrorMessage
		}

		// Log the actual error here, but only if it contains a message.
		if retErr.Error() != "" {
			c.logger.Printf("call request: id='%v'; returned error: %v", header.ID, retErr)
		}

		// Call the error hook, if defined.
		if c.errorHook != nil {
			c.errorHook(c, header.ID, retErr)
		}
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
func (c *Control) handleCallReturn(headerData, payloadData []byte) (err error) {
	// Decode the header.
	var header api.ControlReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode call return: %v", err)
	}

	// Get the channel by the key.
	channel := c.callRetChain.get(header.Key)
	if channel == nil {
		return fmt.Errorf("return request failed: no return channel set (call timeout exceeded?)")
	}

	// Create a new context.
	ctx := newContext(c, payloadData)

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
