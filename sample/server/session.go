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
	"log"
	"net"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/events"
	"github.com/desertbit/orbit/sample/api"
)

type Session struct {
	*orbit.Session
}

func newSession(orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	s.OnNewStream(api.ChannelIDRaw, handleStreamRaw)
	s.OnNewStream(api.ChannelIDPacket, handleStreamPacket)

	s.OnNewStream(api.ChannelIDEvent, func(stream net.Conn) error {
		evs := events.New(stream, nil)
		l := evs.OnEvent(api.HelloEvent)
		data := <-l.C
		fmt.Println(data.Data)
		return nil
	})

	// Signalize the session that initialization is done.
	// Start accepting incoming channel streams.
	s.Ready()

	// Log if the session closes.
	s.OnClose(func() error {
		log.Println("session closed")
		return nil
	})

	return
}
