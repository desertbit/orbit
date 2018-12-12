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
	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/signaler"
)

type Session struct {
	*orbit.Session

	server *Server

	ctrl *control.Control
	sig  *signaler.Signaler
}

func newSession(server *Server, orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
		server:  server,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Log if the session closes.
	s.OnClose(func() error {
		return nil
	})

	s.ctrl, s.sig, err = s.Init(&orbit.Init{
		AcceptStreams: orbit.InitAcceptStreams{
			api.ChannelOrbit: handleStreamOrbit,
		},
		Control: orbit.InitControl{
			Funcs: control.Funcs{
				api.ControlServerInfo: s.serverInfo,
			},
		},
		Signaler: orbit.InitSignaler{
			Config: nil,
			Signals: []orbit.InitSignal{
				{
					ID: api.SignalTimeBomb,
				},
				{
					ID:     api.SignalNewsletter,
					Filter: s.newsletterFilter,
				},
				{
					ID:     api.SignalChatIncomingMessage,
					Filter: s.chatFilter,
				},
			},
		},
	})
	if err != nil {
		return
	}

	_ = s.sig.OnSignalFunc(api.SignalChatSendMessage, s.onChatSendMessage)

	s.ctrl.Ready()
	s.sig.Ready()

	return
}
