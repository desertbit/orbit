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

package main

import (
	"fmt"
	"net"

	"github.com/desertbit/orbit"
)

const (
	serverHandleNewSessionRoutines = 3
)

type Server struct {
	*orbit.Server
}

func NewServer(listenAddr string) (s *Server, err error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return
	}

	s = &Server{
		Server: orbit.NewServer(ln, nil),
	}

	// TODO:
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
			println("new session")

			_, err := newSession(session)
			if err != nil {
				// TODO:
				fmt.Println(err)
			}
		}
	}
}
