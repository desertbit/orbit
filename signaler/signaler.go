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

package signaler

import (
	"log"
	"net"
	"sync"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/control"

	"github.com/desertbit/closer"
)

const (
	cmdSetSignal       = "SetSignal"
	cmdTriggerSignal   = "TriggerSignal"
	cmdSetSignalFilter = "SetSignalFilter"
)

type Signaler struct {
	closer.Closer

	ctrl   *control.Control
	codec  codec.Codec
	logger *log.Logger

	signalsMutex sync.Mutex
	signals      map[string]*signal

	lsMapMutex sync.Mutex
	lsMap      map[string]*listeners
}

func New(conn net.Conn, config *control.Config) (s *Signaler) {
	ctrl := control.New(conn, config)
	s = &Signaler{
		Closer:  ctrl,
		ctrl:    ctrl,
		codec:   ctrl.Codec(),
		logger:  ctrl.Logger(),
		signals: make(map[string]*signal),
		lsMap:   make(map[string]*listeners),
	}

	s.ctrl.AddFuncs(control.Funcs{
		cmdSetSignal:       s.setSignal,
		cmdTriggerSignal:   s.triggerSignal,
		cmdSetSignalFilter: s.setSignalFilter,
	})
	return
}

// Ready signalizes that the initialization is done.
// Signaler can now be triggered.
// This should be only called once.
func (s *Signaler) Ready() {
	s.ctrl.Ready()
}
