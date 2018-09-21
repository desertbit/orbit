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
	handleNewStreamRoutines = 2

	openStreamWriteTimeout    = 7 * time.Second
	acceptStreamReadTimeout   = 7 * time.Second
	acceptStreamMaxHeaderSize = 5 * 1024 // 5 KB
)

type Session struct {
	closer.Closer

	// Value is a custom value which can be set.
	Value interface{}

	config *Config
	logger *log.Logger
	conn   net.Conn
	ys     *yamux.Session
	id     string

	newStreamChan chan net.Conn

	channelMapMutex sync.Mutex
	channelMap      map[string]Channel
}

func newSession(
	conn net.Conn,
	ys *yamux.Session,
	config *Config,
) (s *Session) {
	s = &Session{
		Closer:        closer.New(),
		config:        config,
		logger:        config.Logger,
		conn:          conn,
		ys:            ys,
		newStreamChan: make(chan net.Conn, 2),
		channelMap:    make(map[string]Channel),
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

// Ready signalizes the session that the initialization is done.
// The session starts accepting new incoming channel streams.
// This should be only called once per session!
func (s *Session) Ready() {
	// Start routines.
	go s.acceptStreamRoutine()
	for i := 0; i < handleNewStreamRoutines; i++ {
		go s.handleNewStreamRoutine()
	}
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

// LocalAddr returns the local network address.
func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// AddChannels registers new channels.
// This method is thread-safe.
func (s *Session) AddChannels(channels Channels) {
	s.channelMapMutex.Lock()
	for _, c := range channels {
		s.channelMap[c.ID()] = c
	}
	s.channelMapMutex.Unlock()

	// Call the init method.
	for _, c := range channels {
		c.Init(s)
	}
}

// AddChannel registers a single new channel.
// This method is thread-safe.
func (s *Session) AddChannel(c Channel) {
	s.channelMapMutex.Lock()
	s.channelMap[c.ID()] = c
	s.channelMapMutex.Unlock()

	// Call the init method.
	c.Init(s)
}

// Channel returns the channel specified by its ID.
// Returns an error if the channel does not exists.
func (s *Session) Channel(id string) (c Channel, err error) {
	s.channelMapMutex.Lock()
	c = s.channelMap[id]
	s.channelMapMutex.Unlock()

	if c == nil {
		err = fmt.Errorf("channel does not exists: id='%v'", id)
	}
	return
}

// OpenStream opens a new stream with the given channel ID.
func (s *Session) OpenStream(id string) (stream net.Conn, err error) {
	stream, err = s.ys.Open()
	if err != nil {
		return
	}

	data := api.InitStream{
		ID: id,
	}

	err = packet.WriteEncode(&data, msgpack.Codec, stream, acceptStreamMaxHeaderSize, openStreamWriteTimeout)
	if err != nil {
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

func (s *Session) acceptStreamRoutine() {
	defer s.Close()

	for {
		if s.IsClosed() {
			return
		}

		stream, err := s.ys.Accept()
		if err != nil {
			if !s.IsClosed() && err != io.EOF {
				s.logger.Printf("session: failed to accept stream: %v\n", err)
			}
			return
		}

		select {
		case s.newStreamChan <- stream:
		case <-s.CloseChan():
			return
		}
	}
}

func (s *Session) handleNewStreamRoutine() {
	defer s.Close()

	sessionCloseChan := s.CloseChan()

	for {
		select {
		case <-sessionCloseChan:
			return

		case stream := <-s.newStreamChan:
			err := s.handleNewStream(stream)
			if err != nil {
				s.logger.Printf("session: failed to handle new stream: %v\n", err)
			}
		}
	}
}

func (s *Session) handleNewStream(stream net.Conn) (err error) {
	// Close the stream on error.
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()

	// Catch panics. Might be caused by the channel interface.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Decode the header data.
	var data api.InitStream
	err = packet.ReadDecode(&data, msgpack.Codec, stream, acceptStreamMaxHeaderSize, acceptStreamReadTimeout)
	if err != nil {
		return fmt.Errorf("init stream header: %v", err)
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Obtain the channel and handle the new stream.
	c, err := s.Channel(data.ID)
	if err != nil {
		return
	}
	err = c.AcceptStream(stream)
	if err != nil {
		return fmt.Errorf("id='%v': %v", data.ID, err)
	}

	return
}
