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
	"time"

	"github.com/desertbit/orbit/examples/full/api"
	"github.com/desertbit/orbit/pkg/control"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/signaler"
)

type Session struct {
	*orbit.Session

	ctrl *control.Control
	sig  *signaler.Signaler

	uptime time.Time
}

func newSession(orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
		uptime:  time.Now(),
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	s.ctrl, s.sig, err = s.Init(&orbit.Init{
		Control: orbit.InitControl{
			Funcs: map[string]control.Func{
				api.ControlClientInfo: s.clientsInfo,
			},
		},
		Signaler: orbit.InitSignaler{
			Signals: []orbit.InitSignal{
				{
					ID: api.SignalChatSendMessage,
				},
			},
		},
	})
	if err != nil {
		return
	}

	s.setupSignals()

	s.ctrl.Ready()
	s.sig.Ready()

	return
}

func (s *Session) setupSignals() {
	_ = s.sig.OnSignalFunc(api.SignalTimeBomb, s.onEventTimeBomb)
	_ = s.sig.OnSignalFunc(api.SignalNewsletter, s.onNewsletter)
	_ = s.sig.OnSignalFunc(api.SignalChatIncomingMessage, s.onChatIncomingMessage)
}
