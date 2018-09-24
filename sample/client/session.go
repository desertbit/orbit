/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
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
	"github.com/desertbit/orbit/roc"
	"github.com/desertbit/orbit/sample/api"
	"net"
	"time"

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

	calls, events, err := s.Init(&orbit.Init{
		ROC: orbit.InitROC{
			Funcs:  map[string]roc.Func{},
			Config: nil, // Optional. Can be removed from here...
		},
		ROE: orbit.InitROE{
			Events: []orbit.InitEvent{
				{
					ID:     api.EventFilter,
					Filter: filter,
				},
			},
		},
	})
	if err != nil {
		return
	}

	calls.Ready()
	events.Ready()

	time.Sleep(time.Second)

	events.TriggerEvent(api.EventFilter, &api.EventData{ID: "5", Name: "Hello"})

	time.Sleep(time.Second)
	s.Close()

	return
}
