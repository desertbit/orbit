/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers@desertbit.com>
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
	"log"
	"net"
	"sync"

	"github.com/desertbit/orbit/internal/utils"

	"github.com/desertbit/closer"
)

const (
	sessionIDLength = 20
	serverWorkers   = 5

	newConnChanSize    = 5
	newSessionChanSize = 5
)

// Server implements a simple orbit server.
type Server struct {
	closer.Closer

	ln     net.Listener
	logger *log.Logger
	config *Config

	sessionsMutex sync.RWMutex
	sessions      map[string]*Session

	newConnChan    chan net.Conn
	newSessionChan chan *Session
}

// NewServer creates a new orbit listener server.
func NewServer(ln net.Listener, config *Config) *Server {
	config = prepareConfig(config)

	l := &Server{
		Closer:         closer.New(),
		ln:             ln,
		logger:         config.Logger,
		config:         config,
		sessions:       make(map[string]*Session),
		newConnChan:    make(chan net.Conn, newConnChanSize),
		newSessionChan: make(chan *Session, newSessionChanSize),
	}
	l.OnClose(ln.Close)

	for w := 0; w < serverWorkers; w++ {
		go l.handleConnectionLoop()
	}

	return l
}

// Listen for new socket connections.
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

// Sessions returns a list of all current connected sessions.
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

func (l *Server) handleConnectionLoop() {
	closeChan := l.CloseChan()

	for {
		select {
		case <-closeChan:
			return

		case conn := <-l.newConnChan:
			err := l.handleConnection(conn)
			if err != nil {
				l.logger.Printf("server: handle new connection: %v\n", err)
			}
		}
	}
}

func (l *Server) handleConnection(conn net.Conn) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("catched panic: %v", e)
		}
	}()

	s, err := ServerSession(conn, l.config)
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

		// Wait for the session to close.
		select {
		case <-l.CloseChan():
		case <-s.CloseChan():
		}

		l.sessionsMutex.Lock()
		delete(l.sessions, id)
		l.sessionsMutex.Unlock()
	}()

	// Finally pass the new seession to the channel.
	l.newSessionChan <- s

	return
}
