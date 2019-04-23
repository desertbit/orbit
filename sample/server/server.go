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

package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/sample/auth"

	"github.com/desertbit/orbit"
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
		closingChan    = s.ClosingChan()
		newSessionChan = s.NewSessionChan()
	)

	for {
		select {
		case <-closingChan:
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
