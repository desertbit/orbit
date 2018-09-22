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
	"net"
	"time"

	"github.com/desertbit/orbit/sample/api"

	"github.com/desertbit/orbit"
)

type Session struct {
	*orbit.Session
}

func NewSession(remoteAddr string) (s *Session, err error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}

	orbitSession, err := orbit.ClientSession(conn, nil)
	if err != nil {
		return
	}

	s = &Session{
		Session: orbitSession,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Open a new custom channel stream to the peer.
	stream, err := s.OpenStream(api.ChannelIDOrbit)
	if err != nil {
		return
	}
	go streamOrbitRoutine(stream)

	// Signalize the session that initialization is done.
	// Start accepting incoming channel streams.
	s.Ready()

	// TODO:
	defer s.Close()
	time.Sleep(time.Second)

	return
}
