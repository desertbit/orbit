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

package roc

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
	defaultErrorMessage = "method call failed"

	typeCall       byte = 0
	typeCallReturn byte = 1
)

var (
	// ErrClosed defines the error if the socket connection is closed.
	ErrClosed = errors.New("socket closed")

	// ErrTimeout defines the error if the call timeout is reached.
	ErrTimeout = errors.New("timeout")
)

// Func defines a callable PAKT function.
type Func func(ctx *Context) (data interface{}, err error)

// Funcs defines a set of functions.
type Funcs map[string]Func

// CallHook defines the callback function.
type CallHook func(c *ROC, funcID string, ctx *Context)

// ErrorHook defines the error callback function.
type ErrorHook func(c *ROC, funcID string, err error)

// ROC TODO.
type ROC struct {
	closer.Closer

	config     *Config
	logger     *log.Logger
	conn       net.Conn
	writeMutex sync.Mutex

	funcChain    *chain
	funcMapMutex sync.RWMutex
	funcMap      map[string]Func

	callHook  CallHook
	errorHook ErrorHook
}

// New creates a new PAKT socket using the passed connection.
// One variadic argument specifies the socket ID.
// Ready() must be called to start the socket read routine.
func New(conn net.Conn, config *Config) *ROC {
	// Create a new socket.
	config = prepareConfig(config)
	r := &ROC{
		Closer:    closer.New(),
		config:    config,
		logger:    config.Logger,
		conn:      conn,
		funcChain: newChain(),
		funcMap:   make(map[string]Func),
	}
	r.OnClose(conn.Close)

	return r
}

func (r *ROC) Logger() *log.Logger {
	return r.logger
}

func (r *ROC) Codec() codec.Codec {
	return r.config.Codec
}

// Ready signalizes the ROC that the initialization is done.
// The socket starts reading from the underlying connection.
// This should be only called once per socket.
func (r *ROC) Ready() {
	// Start the service routines.
	go r.readRoutine()
}

// LocalAddr returns the local network address.
func (r *ROC) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (r *ROC) RemoteAddr() net.Addr {
	return r.conn.RemoteAddr()
}

// SetCallHook sets the call hook function which is triggered, if a local
// remote callable function will be called. This hook can be used for logging purpose.
// Only set this hook during initialization.
func (r *ROC) SetCallHook(h CallHook) {
	r.callHook = h
}

// SetErrorHook sets the error hook function which is triggered, if a local
// remote callable function returns an error. This hook can be used for logging purpose.
// Only set this hook during initialization.
func (r *ROC) SetErrorHook(h ErrorHook) {
	r.errorHook = h
}

// AddFunc registers a remote function.
// This method is thread-safe.
func (r *ROC) AddFunc(id string, f Func) {
	r.funcMapMutex.Lock()
	r.funcMap[id] = f
	r.funcMapMutex.Unlock()
}

// AddFuncs registers a map of remote functions.
// This method is thread-safe.
func (r *ROC) AddFuncs(funcs Funcs) {
	// Lock the mutex.
	r.funcMapMutex.Lock()
	defer r.funcMapMutex.Unlock()

	// Iterate through the map and register the functions.
	for id, f := range funcs {
		r.funcMap[id] = f
	}
}

// OneShot calls a remote function without a return request.
func (r *ROC) OneShot(id string, data interface{}) error {
	header := &api.Call{
		ID: id,
	}

	return r.write(typeCall, header, data)
}

// Call a remote function and wait for its result.
// Pass a byte slice to skip encoding.
// This method blocks until the remote socket function returns.
// Returns ErrTimeout on a timeout.
// Returns ErrClosed if the connection is closed.
// This method is thread-safe.
func (r *ROC) Call(id string, data interface{}) (*Context, error) {
	return r.CallTimeout(id, data, r.config.CallTimeout)
}

// CallTimeout sames as Call but with custom timeout.
func (r *ROC) CallTimeout(id string, data interface{}, timeout time.Duration) (ctx *Context, err error) {
	// Create a new channel with its key.
	key, channel, err := r.funcChain.New()
	if err != nil {
		return
	}
	defer r.funcChain.Delete(key)

	// Create the header.
	header := &api.Call{
		ID:  id,
		Key: key,
	}

	// Write to the client.
	err = r.write(typeCall, header, data)
	if err != nil {
		return
	}

	// Create the timeout.
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	// Wait for a response.
	select {
	case <-r.CloseChan():
		err = ErrClosed

	case <-timeoutTimer.C:
		err = ErrTimeout

	case rDataI := <-channel:
		// Assert the return data.
		rData, ok := rDataI.(retChainData)
		if !ok {
			return nil, fmt.Errorf("failed to assert return data")
		}

		ctx = rData.Context
		err = rData.Err
	}
	return
}

//###############//
//### Private ###//
//###############//

type retChainData struct {
	Context *Context
	Err     error
}

func (r *ROC) write(reqType byte, headerI interface{}, dataI interface{}) (err error) {
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
			payload, err = r.config.Codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("encode: %v", err)
			}
		}
	}

	r.writeMutex.Lock()
	defer r.writeMutex.Unlock()

	err = r.conn.SetWriteDeadline(time.Now().Add(r.config.WriteTimeout))
	if err != nil {
		return err
	}

	_, err = r.conn.Write([]byte{reqType})
	if err != nil {
		return err
	}

	err = packet.Write(r.conn, header, r.config.MaxMessageSize)
	if err != nil {
		return err
	}

	err = packet.Write(r.conn, payload, r.config.MaxMessageSize)
	if err != nil {
		return err
	}

	return nil
}

func (r *ROC) readRoutine() {
	// Close the roc on exit.
	defer r.Close()

	// Warning: don't shadow the error.
	// Otherwise the defered logging won't work!
	var (
		err                     error
		n, bytesRead            int
		reqType                 byte
		reqTypeBuf              = make([]byte, 1)
		headerData, payloadData []byte
	)

	// Catch panics. and log error messages.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}

		// Only log if not closed.
		if err != nil && err != io.EOF && !r.IsClosed() {
			r.logger.Printf("roc: read: %v", err)
		}
	}()

	for {
		bytesRead = 0

		// No timeout, as we need to wait here for any incoming request.
		err = r.conn.SetReadDeadline(time.Time{})
		if err != nil {
			return
		}

		// Read the reqType from the stream.
		// Read in a loop, as Read could potentially return 0 read bytes.
		for bytesRead == 0 {
			n, err = r.conn.Read(reqTypeBuf)
			if err != nil {
				return
			}
			bytesRead += n
		}

		reqType = reqTypeBuf[0]

		err = r.conn.SetReadDeadline(time.Now().Add(r.config.ReadTimeout))
		if err != nil {
			return
		}

		// Read the header from the stream.
		headerData, err = packet.Read(r.conn, nil, r.config.MaxMessageSize)
		if err != nil {
			return
		}

		// Read the payload from the stream.
		payloadData, err = packet.Read(r.conn, nil, r.config.MaxMessageSize)
		if err != nil {
			return
		}

		// Handle the received message in a new goroutine.
		go func() {
			gerr := r.handleReceivedMessage(reqType, headerData, payloadData)
			if gerr != nil {
				r.logger.Printf("roc: handleReceivedMessage: %v", gerr)
			}
		}()
	}
}

func (r *ROC) handleReceivedMessage(reqType byte, headerData, payloadData []byte) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Check the request type.
	switch reqType {
	case typeCall:
		err = r.handleCallRequest(headerData, payloadData)

	case typeCallReturn:
		err = r.handleCallReturnRequest(headerData, payloadData)

	default:
		err = fmt.Errorf("invalid request type: %v", reqType)
	}
	return
}

func (r *ROC) handleCallRequest(headerData, payloadData []byte) (err error) {
	var header api.Call
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode roc call: %v", err)
	}

	r.funcMapMutex.RLock()
	f, ok := r.funcMap[header.ID]
	r.funcMapMutex.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exist: id=%v", header.ID)
	}

	ctx := newContext(r, payloadData)

	// Call the call hook if defined.
	if r.callHook != nil {
		r.callHook(r, header.ID, ctx)
	}

	var (
		code int
		msg  string
	)

	retData, retErr := f(ctx)
	if retErr != nil {
		r.logger.Printf("call request: id='%v': returned error: %v", header.ID, retErr)

		// Decide what to send back to the caller.
		if cErr, ok := retErr.(Error); ok {
			code = cErr.Code()
			msg = cErr.Msg()
		} else if r.config.SendErrToCaller {
			msg = retErr.Error()
		}

		// Ensure an error message is always set.
		if msg == "" {
			msg = defaultErrorMessage
		}
	}

	// Skip the return if this is a oneshot call.
	if header.Key == "" {
		return nil
	}

	retHeader := &api.CallReturn{
		Key:  header.Key,
		Msg:  msg,
		Code: code,
	}

	err = r.write(typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("call request: send return request: %v", err)
	}

	// Call the error hook if defined.
	if retErr != nil && r.errorHook != nil {
		r.errorHook(r, header.ID, retErr)
	}

	return nil
}

func (r *ROC) handleCallReturnRequest(headerData, payloadData []byte) (err error) {
	var header api.CallReturn
	err = msgpack.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode call return: %v", err)
	}

	// Get the channel by the key.
	channel := r.funcChain.Get(header.Key)
	if channel == nil {
		return fmt.Errorf("call return request failed (call timeout exceeded?)")
	}

	// Create a new context.
	ctx := newContext(r, payloadData)

	// Create the channel data.
	rData := retChainData{Context: ctx}

	// Create a roc.Error, if an error is present.
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
			return fmt.Errorf("call return request failed (call timeout exceeded?)")
		}
	}
}
