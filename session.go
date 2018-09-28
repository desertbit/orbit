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

package orbit

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/packet"

	"github.com/desertbit/closer"
	"github.com/hashicorp/yamux"
)

const (
	openStreamWriteTimeout    = 7 * time.Second
	acceptStreamReadTimeout   = 7 * time.Second
	acceptStreamMaxHeaderSize = 5 * 1024 // 5 KB
)

type AuthFunc func(net.Conn) (interface{}, error)

type AcceptStreamFunc func(net.Conn) error

type Session struct {
	closer.Closer

	// Value is a custom value which can be set.
	Value interface{}

	config   *Config
	logger   *log.Logger
	conn     net.Conn
	ys       *yamux.Session
	isClient bool
	id       string

	acceptStreamFuncsMutex sync.Mutex
	acceptStreamFuncs      map[string]AcceptStreamFunc
}

func newSession(
	conn net.Conn,
	ys *yamux.Session,
	config *Config,
	isClient bool,
) (s *Session) {
	s = &Session{
		Closer:            closer.New(),
		config:            config,
		logger:            config.Logger,
		conn:              conn,
		ys:                ys,
		isClient:          isClient,
		acceptStreamFuncs: make(map[string]AcceptStreamFunc),
	}
	s.OnClose(conn.Close)
	s.OnClose(ys.Close)

	// Close if the underlying connection closes,
	go func() {
		select {
		case <-s.CloseChan():
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

// IsClient returns a boolean whenever this session is a client connection.
func (s *Session) IsClient() bool {
	return s.isClient
}

// IsServer returns a boolean whenever this session is a server connection.
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

// OpenStream opens a new stream with the given channel ID.
func (s *Session) OpenStream(channel string) (stream net.Conn, err error) {
	return s.OpenStreamTimeout(channel, openStreamWriteTimeout)
}

// OpenStreamTimeout opens a new stream with the given channel ID.
// Expires after the timeout and returns ErrTimeout.
func (s *Session) OpenStreamTimeout(channel string, timeout time.Duration) (stream net.Conn, err error) {
	stream, err = s.ys.Open()
	if err != nil {
		return
	}

	data := api.InitStream{
		Channel: channel,
	}

	err = packet.WriteEncode(stream, &data, msgpack.Codec, acceptStreamMaxHeaderSize, timeout)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			err = ErrTimeout
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

func (s *Session) acceptStreamRoutine() {
	defer s.Close()

	for {
		if s.IsClosed() {
			return
		}

		stream, err := s.ys.Accept()
		if err != nil {
			if !s.IsClosed() && err != io.EOF {
				s.logger.Printf("session: failed to accept stream: %v", err)
			}
			return
		}

		// Run this in a new goroutine.
		go func() {
			gerr := s.handleNewStream(stream)
			if gerr != nil {
				s.logger.Printf("session: failed to handle new stream: %v", gerr)
			}
		}()
	}
}

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

	// Decode the header data.
	var data api.InitStream
	err = packet.ReadDecode(stream, &data, msgpack.Codec, acceptStreamMaxHeaderSize, acceptStreamReadTimeout)
	if err != nil {
		return fmt.Errorf("init stream header: %v", err)
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Obtain the channel and handle the new stream.
	f, err := s.getAcceptStreamFunc(data.Channel)
	if err != nil {
		return
	}
	err = f(stream)
	if err != nil {
		return fmt.Errorf("channel='%v': %v", data.Channel, err)
	}

	return
}

func (s *Session) getAcceptStreamFunc(channel string) (f AcceptStreamFunc, err error) {
	s.acceptStreamFuncsMutex.Lock()
	f = s.acceptStreamFuncs[channel]
	s.acceptStreamFuncsMutex.Unlock()

	if f == nil {
		err = fmt.Errorf("channel does not exists: '%v'", channel)
	}
	return
}
