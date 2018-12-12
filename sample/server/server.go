/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
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

package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/sample/auth"
	"github.com/desertbit/orbit/signaler"
)

const (
	serverHandleNewSessionRoutines = 3
)

type Server struct {
	*orbit.Server

	uptime time.Time

	sessionsMutex sync.RWMutex
	sessions      map[string]*Session

	chatSigGroup *signaler.Group
}

func NewServer(listenAddr string, authHook auth.GetHashHook) (s *Server, err error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return
	}

	s = &Server{
		Server: orbit.NewServer(ln, &orbit.ServerConfig{
			Config: &orbit.Config{
				AuthFunc: auth.Server(authHook),
			},
		}),
		uptime:       time.Now(),
		sessions:     make(map[string]*Session),
		chatSigGroup: signaler.NewGroup(),
	}

	// Always close the server on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	for i := 0; i < serverHandleNewSessionRoutines; i++ {
		go s.handleNewSessionRoutine()
	}

	return
}

func (s *Server) handleNewSessionRoutine() {
	defer s.Close()

	var (
		serverCloseChan = s.CloseChan()
		newSessionChan  = s.NewSessionChan()
	)

	for {
		select {
		case <-serverCloseChan:
			return

		case session := <-newSessionChan:
			sess, err := newSession(s, session)
			if err != nil {
				fmt.Printf("handleNewSessionRoutine: %v\n", err)
			}

			s.addSession(sess)

			// Add the signaler of it to our signaler group.
			s.chatSigGroup.Add(sess.sig)

			// Once the session closes, remove it from the sessions map.
			sess.OnClose(func() error {
				s.removeSession(sess)
				return nil
			})
		}
	}
}

func (s *Server) addSession(session *Session) {
	s.sessionsMutex.Lock()
	s.sessions[session.ID()] = session
	s.sessionsMutex.Unlock()
}

func (s *Server) removeSession(session *Session) {
	s.sessionsMutex.Lock()
	delete(s.sessions, session.ID())
	s.sessionsMutex.Unlock()
}

func (s *Server) Sessions() (sessions []*Session) {
	s.sessionsMutex.RLock()
	sessions = make([]*Session, len(s.sessions))
	i := 0
	for _, sn := range s.sessions {
		sessions[i] = sn
		i++
	}
	s.sessionsMutex.RUnlock()
	return sessions
}
