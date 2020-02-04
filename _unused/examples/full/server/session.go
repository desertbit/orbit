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
	"github.com/desertbit/orbit/examples/full/api"
	"github.com/desertbit/orbit/pkg/control"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/signaler"
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

	s.ctrl, s.sig, err = s.Init(&orbit.SessionHandler{
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
