/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
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
offers a RPC-like request/response mechanism.

uses packet format
automatic encoding
a-/synchronous requests
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

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/packet"

	"github.com/desertbit/closer"
)

const (
	// A default error message that is sent back to a caller of remote function,
	// in case the Func of the remote did not return an error that conforms to
	// our Error interface.
	// This is done to prevent sensitive information to leak to the outside that
	// is usually carried in normal errors.
	defaultErrorMessage = "method call failed"

	// The first byte send in a request to indicate a call to a remote function.
	typeCall       byte = 0

	// The first byte send in a request to indicate the response from a remote
	// called function.
	typeCallReturn byte = 1
)

var (
	// ErrClosed defines the error if the socket connection is closed.
	ErrClosed = errors.New("socket closed")

	// ErrTimeout defines the error if the call timeout is reached.
	ErrTimeout = errors.New("timeout")
)

// The Func type defines a callable Control function.
type Func func(ctx *Context) (data interface{}, err error)

// The Funcs type defines a set of callable Control functions.
type Funcs map[string]Func

// The CallHook type defines the callback function that is executed for every
// Func called by the remote peer.
type CallHook func(c *Control, funcID string, ctx *Context)

// The ErrorHook type defines the error callback function that is executed
// for every Func calley by the remote peer, that produces an error.
type ErrorHook func(c *Control, funcID string, err error)

// The Control type is the main struct that implements the network protocol
// offered by this package.
type Control struct {
	closer.Closer

	// Stores the configuration for the Control, such as timeouts, codecs, etc.
	config     *Config
	// The logger used to log messages.
	logger     *log.Logger

	// Synchronises write operations to the network connection.
	connWriteMutex sync.Mutex
	// The raw network connection that is used to send the requests over.
	conn           net.Conn


	callRetChain *chain
	funcMapMutex sync.RWMutex
	funcMap      map[string]Func

	callHook  CallHook
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
// Returns ErrTimeout on a timeout.
// Returns ErrClosed if the connection is closed.
//
// This method is thread-safe.
func (c *Control) CallTimeout(id string, data interface{}, timeout time.Duration) (ctx *Context, err error) {
	// Create a new channel with its key.
	key, channel, err := c.callRetChain.New()
	if err != nil {
		return
	}
	defer c.callRetChain.Delete(key)

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

	// Create the timeout.
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	// Wait for a response.
	select {
	case <-c.CloseChan():
		// Abort if the Control closes
		err = ErrClosed

	case <-timeoutTimer.C:
		// Abort if the deadline is over.
		err = ErrTimeout

	case rDataI := <-channel:
		// Assert the return data.
		rData, ok := rDataI.(chainData)
		if !ok {
			return nil, fmt.Errorf("failed to assert return data")
		}

		ctx = rData.Context
		err = rData.Err
	}
	return
}

// CallAsync calls a remote function in an asynchronous fashion,
// as it does not wait for a response of the peer.
func (c *Control) CallAsync(id string, data interface{}) error {
	// Create the header, but without a key as we do not register a response
	// handler for it.
	header := &api.ControlCall{
		ID: id,
	}

	// Send the request.
	return c.write(typeCall, header, data)
}

//###############//
//### Private ###//
//###############//

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

	c.connWriteMutex.Lock()
	defer c.connWriteMutex.Unlock()

	err = c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte{reqType})
	if err != nil {
		return err
	}

	err = packet.Write(c.conn, header, c.config.MaxMessageSize)
	if err != nil {
		return err
	}

	err = packet.Write(c.conn, payload, c.config.MaxMessageSize)
	if err != nil {
		return err
	}

	return nil
}

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

	// Catch panics. and log error messages.
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
			gerr := c.handleReceivedMessage(reqType, headerData, payloadData)
			if gerr != nil {
				c.logger.Printf("control: handleReceivedMessage: %v", gerr)
			}
		}()
	}
}

func (c *Control) handleReceivedMessage(reqType byte, headerData, payloadData []byte) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = c.handleCallRequest(headerData, payloadData)

	case typeCallReturn:
		err = c.handleReturnRequest(headerData, payloadData)

	default:
		err = fmt.Errorf("invalid request type: %v", reqType)
	}
	return
}

func (c *Control) handleCallRequest(headerData, payloadData []byte) (err error) {
	var header api.ControlCall
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode control call: %v", err)
	}

	c.funcMapMutex.RLock()
	f, ok := c.funcMap[header.ID]
	c.funcMapMutex.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exist: id=%v", header.ID)
	}

	ctx := newContext(c, payloadData)

	// Call the call hook if defined.
	if c.callHook != nil {
		c.callHook(c, header.ID, ctx)
	}

	var (
		code int
		msg  string
	)

	retData, retErr := f(ctx)
	if retErr != nil {
		c.logger.Printf("call request: id='%v': returned error: %v", header.ID, retErr)

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
	}

	// Skip the return if this is a oneshot call.
	if header.Key == 0 {
		return nil
	}

	retHeader := &api.ControlReturn{
		Key:  header.Key,
		Msg:  msg,
		Code: code,
	}

	err = c.write(typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("call request: send return request: %v", err)
	}

	// Call the error hook if defined.
	if retErr != nil && c.errorHook != nil {
		c.errorHook(c, header.ID, retErr)
	}

	return nil
}

func (c *Control) handleReturnRequest(headerData, payloadData []byte) (err error) {
	var header api.ControlReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode call return: %v", err)
	}

	// Get the channel by the key.
	channel := c.callRetChain.Get(header.Key)
	if channel == nil {
		return fmt.Errorf("return request failed: no return channel set (call timeout exceeded?)")
	}

	// Create a new context.
	ctx := newContext(c, payloadData)

	// Create the channel data.
	rData := chainData{Context: ctx}

	// Create a control.Error, if an error is present.
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
