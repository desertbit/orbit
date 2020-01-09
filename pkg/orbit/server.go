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
	"runtime/debug"
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/utils"
	"github.com/rs/zerolog"
)

const (
	// The length of the randomly created session ids.
	sessionIDLength = 20

	// The maximum number of times it is tried to generate a unique random session id.
	maxRetriesGenSessionID = 10
)

type NewSessionCreatedFunc func(s *Session)

// Server implements a simple orbit server. It listens with serverWorkers many
// routines for incoming connections.
type Server struct {
	closer.Closer

	ln  Listener
	cf  *ServerConfig
	log *zerolog.Logger

	sessionsMutex sync.RWMutex
	sessions      map[string]*Session

	newConnChan    chan Conn
	newSessionChan chan *Session

	newSessionCreatedFuncsMx sync.RWMutex
	newSessionCreatedFuncs   []NewSessionCreatedFunc
}

// NewServer creates a new orbit server. A listener is required
// the server will use to listen for incoming connections.
// A config can be provided, where every property of it that has not
// been set will be initialized with a default value.
// That makes it possible to overwrite only the interesting properties
// for the caller.
func NewServer(cl closer.Closer, ln Listener, cf *ServerConfig) *Server {
	return newServer(cl, ln, cf)
}

// newServer is the internal helper to create a new orbit server.
func newServer(cl closer.Closer, ln Listener, cf *ServerConfig) *Server {
	// Prepare the config.
	cf = prepareServerConfig(cf)

	s := &Server{
		Closer:      cl,
		ln:          ln,
		cf:          cf,
		log:         cf.Log,
		sessions:    make(map[string]*Session),
		newConnChan: make(chan Conn, cf.NewConnChanSize),
	}
	s.OnClosing(ln.Close)

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

		s.newConnChan <- conn
	}
}

// Session obtains a session by its ID.
// Returns nil if not found.
func (s *Server) Session(id string) (sn *Session) {
	s.sessionsMutex.RLock()
	sn = s.sessions[id]
	s.sessionsMutex.RUnlock()
	return
}

// Sessions returns a list of all currently connected sessions.
func (s *Server) Sessions() []*Session {
	// Lock the mutex.
	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()

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

func (s *Server) OnNewSessionCreated(f NewSessionCreatedFunc) {
	s.newSessionCreatedFuncsMx.Lock()
	s.newSessionCreatedFuncs = append(s.newSessionCreatedFuncs, f)
	s.newSessionCreatedFuncsMx.Unlock()
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

		case conn := <-s.newConnChan:
			err := s.handleConnection(conn)
			if err != nil {
				s.log.Error().
					Err(err).
					Msg("server: handle new connection")
			}
		}
	}
}

// handleConnection handles one new connection.
// It creates a new server session and stores it in the sessions map.
// It starts a routine that takes care of removing the session from said map
// once it has been closed.
// The session is finally passed to the new session channel.
// This method recovers from panics.
func (s *Server) handleConnection(conn Conn) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: %v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: %v", e)
			}
		}
	}()

	// Create a new server session.
	sn, err := newServerSession(s.CloserOneWay(), conn, s.cf.Config)
	if err != nil {
		return
	}

	// Close the session on error.
	defer func() {
		if err != nil {
			sn.Close_()
		}
	}()

	// Generate a unique random id for the new session.
	var (
		id    string
		added bool
	)
	for i := 0; i < maxRetriesGenSessionID; i++ {
		id, err = utils.RandomString(sessionIDLength)
		if err != nil {
			return
		}

		added = func() bool {
			s.sessionsMutex.Lock()
			defer s.sessionsMutex.Unlock()

			if _, ok := s.sessions[id]; ok {
				return false
			}

			sn.id = id
			s.sessions[id] = sn
			return true
		}()
		if added {
			break
		}
	}
	if !added {
		return fmt.Errorf("failed to generate unique random session id")
	}

	sn.OnClosing(func() error {
		// Speed up the closing process when the server closes by returning directly.
		if s.IsClosing() {
			return nil
		}

		s.sessionsMutex.Lock()
		delete(s.sessions, id)
		s.sessionsMutex.Unlock()
		return nil
	})

	// Session created, call hooks.
	s.newSessionCreatedFuncsMx.RLock()
	funcs := make([]NewSessionCreatedFunc, len(s.newSessionCreatedFuncs))
	copy(funcs, s.newSessionCreatedFuncs)
	s.newSessionCreatedFuncsMx.RUnlock()

	for _, f := range funcs {
		f(sn)
	}

	s.log.Debug().
		Str("ID", id).
		Msg("server: new session")

	return
}
