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
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"

	"github.com/desertbit/closer/v3"
	"github.com/hashicorp/yamux"
)

const (
	// The time duration after which a new opened stream timeouts if the initial
	// data could not be written to the stream.
	// Used on the side that opens the stream.
	openStreamWriteTimeout = 15 * time.Second

	// The time duration after which a new opened stream timeouts if the initial
	// data could not be read from the stream.
	// Used on the side that accepts the stream.
	acceptStreamReadTimeout = 20 * time.Second

	// The maximum size the initial data sent over a new stream may have.
	acceptStreamMaxHeaderSize = 2 * 1024 // 2 KB

	// The timeout for the connection flusher.
	flushTimeout = 7 * time.Second
)

// The AuthFunc type describes the function that is used during the authentication
// phase of the session initialization.
// It may use the given connection to perform some kind of data exchange between
// the client and the server.
// It can return some arbitrary data that will be saved to the session.
// It must return a non nil error, if the authentication did fail.
type AuthFunc func(net.Conn) (value interface{}, err error)

// The AcceptStreamFunc type describes the function that is called whenever
// a new connection is requested on a peer. It must then handle the new
// connection, if it could be set up correctly.
type AcceptStreamFunc func(net.Conn) error

// The Session type describes a orbit session that is used on both the client
// and server side, so in general for peers.
// It contains its underlying connection to the remote peer and may accept
// new incoming connections by defining AcceptStreamFuncs.
type Session struct {
	closer.Closer

	// Value is a custom value which can be set. In case the config contains
	// a valid AuthFunc, the Value will be set to the return value of it.
	Value interface{}

	// The config of this session.
	conf *Config
	// The underlying connection to the remote peer.
	conn net.Conn
	// The underlying yamux session that is used to multiplex the stream.
	ys *yamux.Session
	// A flag whether this session is a client or server session.
	isClient bool
	// The id of the session. The id must be set after the session has been
	// created using SetID().
	id string

	// Synchronises the access to the accept stream functions map.
	acceptStreamFuncsMutex sync.Mutex
	// The functions that have been registered to accept new streams.
	// When a new stream has been opened, the first data sent on the stream must
	// contain the key into this map to retrieve the correct function to handle
	// the stream.
	acceptStreamFuncs map[string]AcceptStreamFunc
}

// newSession creates a new orbit session from the given parameters.
// The created session closes, if the underlying connection or yamux
// session are closed.
func newSession(
	conn net.Conn,
	ys *yamux.Session,
	config *Config,
	isClient bool,
	cl closer.Closer,
) (s *Session) {
	s = &Session{
		Closer:            cl,
		conf:              config,
		conn:              conn,
		ys:                ys,
		isClient:          isClient,
		acceptStreamFuncs: make(map[string]AcceptStreamFunc),
	}
	s.OnClose(conn.Close)
	s.OnClose(ys.Close)

	// Close if the underlying connection or yamux session close.
	go func() {
		select {
		case <-s.ClosingChan():
		case <-ys.CloseChan():
		}
		s.Close()
	}()

	return
}

// ID returns the session ID.
// This must be set manually.
func (s *Session) ID() string {
	return s.id
}

// SetID sets the session ID.
func (s *Session) SetID(id string) {
	s.id = id
}

// IsClient returns whether this session is a client connection.
func (s *Session) IsClient() bool {
	return s.isClient
}

// IsServer returns whether this session is a server connection.
func (s *Session) IsServer() bool {
	return !s.isClient
}

// LocalAddr returns the local network address.
func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// AcceptStream registers the given accept handler for the specific channel.
func (s *Session) AcceptStream(channel string, f AcceptStreamFunc) {
	s.acceptStreamFuncsMutex.Lock()
	s.acceptStreamFuncs[channel] = f
	s.acceptStreamFuncsMutex.Unlock()
}

// OpenStream performs the same task as OpenStreamTimeout, but uses the default
// write timeout openStreamWriteTimeout.
func (s *Session) OpenStream(channel string) (stream net.Conn, err error) {
	return s.OpenStreamTimeout(channel, openStreamWriteTimeout)
}

// OpenStreamTimeout opens a new stream with the given channel ID.
// Expires after the timeout and returns ErrOpenTimeout.
func (s *Session) OpenStreamTimeout(channel string, timeout time.Duration) (stream net.Conn, err error) {
	// Open the stream on the yamux session.
	stream, err = s.ys.Open()
	if err != nil {
		return
	}

	// Create the initial data that signals to the remote peer,
	// which channel we want to open.
	data := api.InitStream{
		Channel: channel,
	}

	// Write the initial request to the stream.
	err = packet.WriteEncode(stream, &data, s.conf.Codec, acceptStreamMaxHeaderSize, timeout)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			err = ErrOpenTimeout
		}
		return
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	return
}

//###############//
//### Private ###//
//###############//

// startRoutines signalizes the session that the initialization is done.
// The session starts accepting new incoming channel streams.
func (s *Session) startRoutines() {
	go s.acceptStreamRoutine()
}

// acceptStreamRoutine is a routine that accepts new connections and
// calls handleNewStream() for each of them.
// Closes together with the session.
func (s *Session) acceptStreamRoutine() {
	defer s.Close()

	for {
		// Quit if the session has been closed.
		if s.IsClosed() {
			return
		}

		// Wait for new incoming connections.
		stream, err := s.ys.Accept()
		if err != nil {
			if !s.IsClosed() && err != io.EOF {
				s.conf.Logger.Printf("session: failed to accept stream: %v", err)
			}
			return
		}

		// Run this in a new goroutine.
		go func() {
			gerr := s.handleNewStream(stream)
			if gerr != nil {
				s.conf.Logger.Printf("session: failed to handle new stream: %v", gerr)
			}
		}()
	}
}

// handleNewStream handles a new incoming stream. It first reads the initial
// stream data from the connection, which tells us the id of the channel that
// should be opened. With this id, we can retrieve the AcceptStreamFunc from the
// session's map and pass it the new stream.
// Recovers from panics.
func (s *Session) handleNewStream(stream net.Conn) (err error) {
	defer func() {
		// Catch panics. Might be caused by the channel interface.
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}

		// Close the stream on error.
		if err != nil {
			stream.Close()
		}
	}()

	// Read the initial data from the stream.
	var data api.InitStream
	err = packet.ReadDecode(
		stream,
		&data,
		s.conf.Codec,
		acceptStreamMaxHeaderSize,
		acceptStreamReadTimeout,
	)
	if err != nil {
		return fmt.Errorf("init stream header: %v", err)
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Obtain the accept stream func.
	f, err := s.getAcceptStreamFunc(data.Channel)
	if err != nil {
		return
	}

	// Pass it the new stream.
	err = f(stream)
	if err != nil {
		return fmt.Errorf("channel='%v': %v", data.Channel, err)
	}

	return
}

// getAcceptStreamFunc retrieves the AcceptStreamFunc from the session's map
// that corresponds to the given channel id.
// This method is thread-safe.
func (s *Session) getAcceptStreamFunc(channel string) (f AcceptStreamFunc, err error) {
	s.acceptStreamFuncsMutex.Lock()
	f = s.acceptStreamFuncs[channel]
	s.acceptStreamFuncsMutex.Unlock()

	if f == nil {
		err = fmt.Errorf("channel does not exists: '%v'", channel)
	}
	return
}
