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
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/signaler"
	"time"

	"github.com/desertbit/orbit"
)

type Session struct {
	*orbit.Session

	ctrl *control.Control
	sig *signaler.Signaler

	uptime time.Time
}

func newSession(orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
		uptime: time.Now(),
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
					ID:     api.SignalFilter,
					Filter: filter,
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
}
