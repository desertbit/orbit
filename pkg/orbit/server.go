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
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/rs/zerolog"
)

type ServerHandler interface {
	SessionHandler

	InitServer(s *Server)
}

// Server implements a simple orbit server.
type Server struct {
	closer.Closer

	ln    Listener
	cf    *ServerConfig
	log   *zerolog.Logger
	h     ServerHandler
	hooks []Hook

	sessionsMx sync.RWMutex
	sessions   map[string]*Session

	connChan    chan Conn
	sessionChan chan *Session
}

// NewServer creates a new orbit server. A listener is required
// the server will use to listen for incoming connections.
// A config can be provided, where every property of it that has not
// been set will be initialized with a default value.
// That makes it possible to overwrite only the interesting properties
// for the caller.
func NewServer(cl closer.Closer, ln Listener, cf *ServerConfig, h ServerHandler, hs ...Hook) *Server {
	return newServer(cl, ln, cf, h, hs...)
}

// newServer is the internal helper to create a new orbit server.
func newServer(cl closer.Closer, ln Listener, cf *ServerConfig, h ServerHandler, hs ...Hook) *Server {
	// Prepare the config.
	cf = prepareServerConfig(cf)

	s := &Server{
		Closer:   cl,
		ln:       ln,
		cf:       cf,
		log:      cf.Log,
		h:        h,
		hooks:    hs,
		sessions: make(map[string]*Session),
		connChan: make(chan Conn, cf.NewConnChanSize),
	}
	s.OnClosing(ln.Close)

	// Call the handler.
	h.InitServer(s)

	// Start the workers that listen for incoming connections.
	for i := 0; i < cf.NewConnNumberWorkers; i++ {
		go s.handleConnectionLoop()
	}

	return s
}

// Listen listens for new socket connections, which it passes to the
// new connection channel that is read by the server workers.
// This method is blocking.
func (s *Server) Listen() error {
	defer s.Close_()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if s.IsClosing() {
				return nil
			}
			return err
		}

		s.connChan <- conn
	}
}

// Session obtains a session by its ID.
// Returns nil if not found.
func (s *Server) Session(id string) (sn *Session) {
	s.sessionsMx.RLock()
	sn = s.sessions[id]
	s.sessionsMx.RUnlock()
	return
}

// Sessions returns a list of all currently connected sessions.
func (s *Server) Sessions() []*Session {
	// Lock the mutex.
	s.sessionsMx.RLock()
	defer s.sessionsMx.RUnlock()

	// Create the slice.
	list := make([]*Session, len(s.sessions))

	// Add all sessions from the map.
	i := 0
	for _, sn := range s.sessions {
		list[i] = sn
		i++
	}

	return list
}

//###############//
//### Private ###//
//###############//

// handleConnectionLoop reads in a loop from the new connection channel
// and calls the handleConnection() function on each read connection.
// Closes, when the server is closed.
func (s *Server) handleConnectionLoop() {
	var (
		closingChan = s.ClosingChan()
	)

	for {
		select {
		case <-closingChan:
			return

		case conn := <-s.connChan:
			s.handleConnection(conn)
		}
	}
}

// handleConnection handles one new connection.
// It creates a new server session and stores it in the sessions map.
// Errors are logged, instead of being returned.
func (s *Server) handleConnection(conn Conn) {
	// Create a new server session.
	sn, err := newServerSession(s.CloserOneWay(), conn, s.cf.Config, s.h, s.hooks)
	if err != nil {
		// Log. Do not use the Err() field, as stack trace formatting is lost then.
		s.log.Error().
			Msgf("server handle new connection: \n%v", err)
		return
	}

	// Save the session in the map.
	var idExists bool
	s.sessionsMx.Lock()
	_, idExists = s.sessions[sn.id]
	if !idExists {
		s.sessions[sn.id] = sn
	}
	s.sessionsMx.Unlock()

	// Close the new session, if its id is already taken.
	if idExists {
		sn.Close_()
		return
	}

	// Remove the session from the session map, once it closes.
	sn.OnClosing(func() error {
		// Speed up the closing process if the server closes.
		if s.IsClosing() {
			return nil
		}

		s.sessionsMx.Lock()
		delete(s.sessions, sn.id)
		s.sessionsMx.Unlock()
		return nil
	})
}
