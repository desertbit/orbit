/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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
package control

import (
	"errors"
	"fmt"
	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/packet"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
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
type Func func(c *Context) (data interface{}, err error)

// Funcs defines a set of functions.
type Funcs map[string]Func

// CallHook defines the callback function.
type CallHook func(s *Control, funcID string, c *Context)

// ErrorHook defines the error callback function.
type ErrorHook func(s *Control, funcID string, err error)

// Control defines the PAKT socket implementation.
type Control struct {
	closer.Closer

	config     *Config
	logger     *log.Logger
	conn       net.Conn
	writeMutex sync.Mutex

	funcMapMutex sync.RWMutex
	funcMap      map[string]Func

	funcChain *chain

	callHook  CallHook
	errorHook ErrorHook
}

// New creates a new PAKT socket using the passed connection.
// One variadic argument specifies the socket ID.
// Ready() must be called to start the socket read routine.
func New(conn net.Conn, config *Config) *Control {
	// Create a new socket.
	config = prepareConfig(config)
	s := &Control{
		Closer:    closer.New(),
		config:    config,
		logger:    config.Logger,
		conn:      conn,
		funcMap:   make(map[string]Func),
		funcChain: newChain(),
	}
	s.OnClose(conn.Close)

	return s
}

// Ready signalizes the Control that the initialization is done.
// The socket starts reading from the underlying connection.
// This should be only called once per socket.
func (s *Control) Ready() {
	// Start the service routines.
	go s.readLoop()
}

// LocalAddr returns the local network address.
func (s *Control) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Control) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// SetCallHook sets the call hook function which is triggered, if a local
// remote callable function will be called. This hook can be used for logging purpose.
// Only set this hook during initialization.
func (s *Control) SetCallHook(h CallHook) {
	s.callHook = h
}

// SetErrorHook sets the error hook function which is triggered, if a local
// remote callable function returns an error. This hook can be used for logging purpose.
// Only set this hook during initialization.
func (s *Control) SetErrorHook(h ErrorHook) {
	s.errorHook = h
}

// RegisterFunc registers a remote function.
// This method is thread-safe.
func (s *Control) RegisterFunc(id string, f Func) {
	s.funcMapMutex.Lock()
	s.funcMap[id] = f
	s.funcMapMutex.Unlock()
}

// RegisterFuncs registers a map of remote functions.
// This method is thread-safe.
func (s *Control) RegisterFuncs(funcs Funcs) {
	// Lock the mutex.
	s.funcMapMutex.Lock()
	defer s.funcMapMutex.Unlock()

	// Iterate through the map and register the functions.
	for id, f := range funcs {
		s.funcMap[id] = f
	}
}

func (s *Control) Call(id string, data interface{}) (*Context, error) {
	return s.CallTimeout(id, data, s.config.CallTimeout)
}

// Call a remote function and wait for its result.
// This method blocks until the remote socket function returns.
// Returns ErrTimeout on a timeout.
// Returns ErrClosed if the connection is closed.
// This method is thread-safe.
func (s *Control) CallTimeout(id string, data interface{}, timeout time.Duration) (ctx *Context, err error) {
	// Create a new channel with its key.
	key, channel, err := s.funcChain.New()
	if err != nil {
		return
	}
	defer s.funcChain.Delete(key)

	// Create the header.
	header := &api.ControlCall{
		FuncID:    id,
		ReturnKey: key,
	}

	// Write to the client.
	err = s.write(typeCall, header, data)
	if err != nil {
		return
	}

	// Create the timeout.
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	// Wait for a response.
	select {
	case <-s.CloseChan():
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

func (s *Control) write(reqType byte, headerI interface{}, dataI interface{}) (err error) {
	var payload, header []byte

	// Marshal the payload data if present.
	if dataI != nil {
		payload, err = s.config.Codec.Encode(dataI)
		if err != nil {
			return fmt.Errorf("encode: %v", err)
		}
	}

	// Marshal the header data if present.
	if headerI != nil {
		header, err = s.config.Codec.Encode(headerI)
		if err != nil {
			return fmt.Errorf("encode header: %v", err)
		}
	}

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	_, err = s.conn.Write([]byte{reqType})
	if err != nil {
		return err
	}

	if len(header) > 0 {
		err = packet.Write(s.conn, s.config.MaxMessageSize, s.config.WriteTimeout, header)
		if err != nil {
			return err
		}
	}

	if len(payload) > 0 {
		err = packet.Write(s.conn, s.config.MaxMessageSize, s.config.WriteTimeout, payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Control) read(buf []byte) (int, error) {
	// Reset the read deadline.
	s.conn.SetReadDeadline(time.Now().Add(readTimeout))

	// Read from the socket connection.
	n, err := s.conn.Read(buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (s *Control) readRoutine() {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			s.logger.Printf("control: read loop: catched panic: %v", e)
		}
	}()

	// Close the socket on exit.
	defer s.Close()

	var (
		bytesRead  int
		reqTypeBuf = make([]byte, 1)
	)

	for {
		// Read the reqType from the stream
		for bytesRead == 0 {
			err := s.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
			if err != nil {
				s.logger.Printf("control: setReadDeadline: %v", err)
				return
			}

			n, err := s.conn.Read(reqTypeBuf)
			if err != nil {
				// Log only if not closed.
				if err != io.EOF && !s.IsClosed() {
					s.logger.Printf("control: read: %v", err)
				}
				return
			}
			bytesRead += n
		}

		reqType := reqTypeBuf[0]

		// Read the header from the stream.
		headerData, err := packet.Read(s.conn, s.config.MaxMessageSize, s.config.ReadTimeout, nil)
		if err != nil {
			s.logger.Printf("control: read: %v", err)
			return
		}

		// Read the payload from the stream.
		payloadData, err := packet.Read(s.conn, s.config.MaxMessageSize, s.config.ReadTimeout, nil)
		if err != nil {
			s.logger.Printf("control: read: %v", err)
			return
		}

		// Handle the received message in a new goroutine.
		go func() {
			err := s.handleReceivedMessage(reqType, headerData, payloadData)
			if err != nil {
				s.logger.Printf("control: handleReceivedMessage: %v", err)
			}
		}()
	}
}

func (s *Control) readLoop() {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			s.logger.Printf("control: read loop: catched panic: %v", e)
		}
	}()

	// Close the socket on exit.
	defer s.Close()

	var err error
	var n, bytesRead int

	// Message Head.
	headBuf := make([]byte, 8)
	var headerLen16 uint16
	var headerLen int
	var payloadLen32 uint32
	var payloadLen int

	// Read loop.
	for {
		// Read the head from the stream.
		bytesRead = 0
		for bytesRead < 8 {
			n, err = s.read(headBuf[bytesRead:])
			if err != nil {
				// Log only if not closed.
				if err != io.EOF && !s.IsClosed() {
					Log.Warningf("socket: read: %v", err)
				}
				return
			}
			bytesRead += n
		}

		// The first byte is the version field.
		// Check if this protocol version matches.
		if headBuf[0] != ProtocolVersion {
			Log.Warningf("socket: read: invalid protocol version: %v != %v", ProtocolVersion, headBuf[0])
			return
		}

		// Extract the head fields.
		reqType := headBuf[1]

		// Extract the header length.
		headerLen16, err = bytesToUint16(headBuf[2:4])
		if err != nil {
			Log.Warningf("socket: read: failed to extract header length: %v", err)
			return
		}
		headerLen = int(headerLen16)

		// Check if the maximum header size is exceeded.
		if headerLen > maxHeaderBufferSize {
			Log.Warningf("socket: read: maximum header size exceeded")
			return
		}

		// Extract the payload length.
		payloadLen32, err = bytesToUint32(headBuf[4:8])
		if err != nil {
			Log.Warningf("socket: read: failed to extract payload length: %v", err)
			return
		}
		payloadLen = int(payloadLen32)

		// Check if the maximum payload size is exceeded.
		if payloadLen > s.maxMessageSize {
			Log.Warningf("socket: read: maximum message size exceeded")
			return
		}

		// Read the header bytes from the stream.
		var headerBuf []byte
		if headerLen > 0 {
			headerBuf = make([]byte, headerLen)
			bytesRead = 0
			for bytesRead < headerLen {
				n, err = s.read(headerBuf[bytesRead:])
				if err != nil {
					// Log only if not closed.
					if err != io.EOF && !s.IsClosed() {
						Log.Warningf("socket: read: %v", err)
					}
					return
				}
				bytesRead += n
			}
		}

		// Read the payload bytes from the stream.
		var payloadBuf []byte
		if payloadLen > 0 {
			payloadBuf = make([]byte, payloadLen)
			bytesRead = 0
			for bytesRead < payloadLen {
				n, err = s.read(payloadBuf[bytesRead:])
				if err != nil {
					// Log only if not closed.
					if err != io.EOF && !s.IsClosed() {
						Log.Warningf("socket: read: %v", err)
					}
					return
				}
				bytesRead += n
			}
		}

		// Handle the received message in a new goroutine.
		go func() {
			err := s.handleReceivedMessage(reqType, headerBuf, payloadBuf)
			if err != nil {
				Log.Warningf("socket: %v", err)
			}
		}()
	}
}

func (s *Control) handleReceivedMessage(reqType byte, headerData, payloadData []byte) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// TODO: Reset call timeout?

	// Check the request type.
	switch reqType {
	case typeCall:
		err = s.handleCallRequest(headerData, payloadData)

	case typeCallReturn:
		err = s.handleCallReturnRequest(headerData, payloadData)

	default:
		err = fmt.Errorf("invalid request type: %v", reqType)
	}
	return
}

func (s *Control) handleReceivedMessage(reqType byte, headerBuf, payloadBuf []byte) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Reset the timeout, because data was successful read from the socket.
	s.resetTimeout()

	// Check the request type.
	switch reqType {
	case typeCall:
		return s.handleCallRequest(headerBuf, payloadBuf)

	case typeCallReturn:
		return s.handleCallReturnRequest(headerBuf, payloadBuf)

	default:
		return fmt.Errorf("invalid request type: %v", reqType)
	}

	return nil
}

func (s *Control) handleCallRequest(headerData, payloadData []byte) (err error) {
	var header api.ControlCall
	err = s.config.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode control call: %v", err)
	}

	s.funcMapMutex.RLock()
	f, ok := s.funcMap[header.FuncID]
	s.funcMapMutex.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exist: id=%v", header.FuncID)
	}

	ctx := newContext(s, payloadData)

	// TODO:
	// Call the call hook if defined.
	if s.callHook != nil {
		s.callHook(s, header.FuncID, ctx)
	}

	retData, retErr := f(ctx)

	var retErrString string
	if retErr != nil {
		retErrString = retErr.Error()
	}

	retHeader := &api.ControlCallReturn{
		ReturnKey: header.ReturnKey,
		ReturnErr: retErrString,
	}

	err = s.write(typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("call request: send return request: %v", err)
	}

	// TODO:
	// Call the error hook if defined.
	if retErr != nil && s.errorHook != nil {
		s.errorHook(s, header.FuncID, retErr)
	}

	return nil
}

func (s *Control) handleCallRequest(headerBuf, payloadBuf []byte) (err error) {
	// Decode the header.
	var header headerCall
	err = s.Codec.Decode(headerBuf, &header)
	if err != nil {
		return fmt.Errorf("decode call header: %v", err)
	}

	// Obtain the function defined by the ID.
	s.funcMapMutex.RLock()
	f, ok := s.funcMap[header.FuncID]
	s.funcMapMutex.RUnlock()
	if !ok {
		return fmt.Errorf("call request: requested function does not exists: id=%v", header.FuncID)
	}

	// Create a new context.
	context := newContext(s, payloadBuf)

	// Call the call hook if defined.
	if s.callHook != nil {
		s.callHook(s, header.FuncID, context)
	}

	// Call the function.
	retData, retErr := f(context)

	// Get the string representation of the error if present.
	var retErrString string
	if retErr != nil {
		retErrString = retErr.Error()
	}

	// Create the return header.
	retHeader := &headerCallReturn{
		ReturnKey: header.ReturnKey,
		ReturnErr: retErrString,
	}

	// Write to the client.
	err = s.write(typeCallReturn, retHeader, retData)
	if err != nil {
		return fmt.Errorf("call request: send return request: %v", err)
	}

	// Call the error hook if defined.
	if retErr != nil && s.errorHook != nil {
		s.errorHook(s, header.FuncID, retErr)
	}

	return nil
}

func (s *Control) handleCallReturnRequest(headerData, payloadData []byte) (err error) {
	var header api.ControlCallReturn
	err = s.config.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("decode control call return: %v", err)
	}

	// Get the channel by the return key.
	channel := s.funcChain.Get(header.ReturnKey)
	if channel == nil {
		return fmt.Errorf("conrol call return request failed (call timeout exceeded?)")
	}

	// Create the error if present.
	var retErr error
	if len(header.ReturnErr) > 0 {
		retErr = errors.New(header.ReturnErr)
	}

	// Create a new context.
	ctx := newContext(s, payloadData)

	// Create the channel data.
	rData := retChainData{
		Context: ctx,
		Err:     retErr,
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
			return fmt.Errorf("control call return request failed (call timeout exceeded?)")
		}
	}
}

func (s *Control) handleCallReturnRequest(headerBuf, payloadBuf []byte) (err error) {
	// Decode the header.
	var header headerCallReturn
	err = s.Codec.Decode(headerBuf, &header)
	if err != nil {
		return fmt.Errorf("decode call return header: %v", err)
	}

	// Get the channel by the return key.
	channel := s.funcChain.Get(header.ReturnKey)
	if channel == nil {
		return fmt.Errorf("call return request failed (call timeout exceeded?)")
	}

	// Create the error if present.
	var retErr error
	if len(header.ReturnErr) > 0 {
		retErr = errors.New(header.ReturnErr)
	}

	// Create a new context.
	context := newContext(s, payloadBuf)

	// Create the channel data.
	rData := retChainData{
		Context: context,
		Err:     retErr,
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
