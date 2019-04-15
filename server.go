/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
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
	"net"
	"sync"

	"github.com/desertbit/orbit/internal/utils"

	"github.com/desertbit/closer"
)

const (
	// The length of the randomly created session ids.
	sessionIDLength = 20
)

// Server implements a simple orbit server. It listens with serverWorkers many
// routines for incoming connections.
type Server struct {
	closer.Closer

	ln   net.Listener
	conf *ServerConfig

	sessionsMutex sync.RWMutex
	sessions      map[string]*Session

	newConnChan    chan net.Conn
	newSessionChan chan *Session
}

// NewServer creates a new orbit server. A listener is required
// the server will use to listen for incoming connections.
// A config can be provided, where every property of it that has not
// been set will be initialized with a default value.
// That makes it possible to overwrite only the interesting properties
// for the caller.
func NewServer(ln net.Listener, config *ServerConfig) *Server {
	return newServer(ln, config, closer.New())
}

// NewServerWithCloser creates a new orbit server just like NewServer() does,
// but you can provide your own closer for it.
func NewServerWithCloser(ln net.Listener, config *ServerConfig, cl closer.Closer) *Server {
	return newServer(ln, config, cl)
}

// newServer is the internal helper to create a new orbit server.
func newServer(ln net.Listener, config *ServerConfig, cl closer.Closer) *Server {
	// Prepare the config.
	config = prepareServerConfig(config)

	l := &Server{
		Closer:         cl,
		ln:             ln,
		conf:           config,
		sessions:       make(map[string]*Session),
		newConnChan:    make(chan net.Conn, config.NewConnChanSize),
		newSessionChan: make(chan *Session, config.NewSessionChanSize),
	}
	l.OnClose(ln.Close)

	// Start the workers that listen for incoming connections.
	for w := 0; w < config.NewConnNumberWorkers; w++ {
		go l.handleConnectionLoop()
	}

	return l
}

// Listen listens for new socket connections, which it passes to the
// new connection channel that is read by the server workers.
// This method is blocking.
func (l *Server) Listen() error {
	defer l.Close()

	for {
		conn, err := l.ln.Accept()
		if err != nil {
			if l.IsClosed() {
				return nil
			}
			return err
		}

		l.newConnChan <- conn
	}
}

// NewSessionChan returns the channel for new incoming sessions.
func (l *Server) NewSessionChan() <-chan *Session {
	return l.newSessionChan
}

// Session obtains a session by its ID.
// Returns nil if not found.
func (l *Server) Session(id string) (s *Session) {
	l.sessionsMutex.RLock()
	s = l.sessions[id]
	l.sessionsMutex.RUnlock()
	return
}

// Sessions returns a list of all currently connected sessions.
func (l *Server) Sessions() []*Session {
	// Lock the mutex.
	l.sessionsMutex.RLock()
	defer l.sessionsMutex.RUnlock()

	// Create the slice.
	list := make([]*Session, len(l.sessions))

	// Add all sessions from the map.
	i := 0
	for _, s := range l.sessions {
		list[i] = s
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
func (l *Server) handleConnectionLoop() {
	closeChan := l.CloseChan()

	for {
		select {
		case <-closeChan:
			return

		case conn := <-l.newConnChan:
			err := l.handleConnection(conn)
			if err != nil {
				l.conf.Logger.Printf("server: handle new connection: %v\n", err)
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
func (l *Server) handleConnection(conn net.Conn) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	// Create a new server session.
	s, err := ServerSession(conn, l.conf.Config)
	if err != nil {
		return
	}

	// Close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Add the new session to the active sessions map.
	// If the ID is already present, then generate a new one.
	var id string
	for {
		id, err = utils.RandomString(sessionIDLength)
		if err != nil {
			return
		}

		added := func() bool {
			l.sessionsMutex.Lock()
			defer l.sessionsMutex.Unlock()

			if _, ok := l.sessions[id]; ok {
				return false
			}

			s.SetID(id)
			l.sessions[id] = s
			return true
		}()
		if added {
			break
		}
	}

	// Remove the session from the active sessions map during close.
	// Also close the session if the server closes.
	go func() {
		defer s.Close()

		// Wait for the session or server to close.
		select {
		case <-l.CloseChan():
			// Speed up the closing process when the server closes
			// by returning directly from here.
			return
		case <-s.CloseChan():
		}

		l.sessionsMutex.Lock()
		delete(l.sessions, id)
		l.sessionsMutex.Unlock()
	}()

	// Finally, pass the new session to the channel.
	l.newSessionChan <- s

	return
}
