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

package events

import (
	"log"
	"net"
	"sync"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/control"

	"github.com/desertbit/closer"
)

const (
	cmdSetEvent       = "SetEvent"
	cmdTriggerEvent   = "TriggerEvent"
	cmdSetEventFilter = "SetEventFilter"
)

type Events struct {
	closer.Closer

	ctrl   *control.Control
	codec  codec.Codec
	logger *log.Logger

	eventMapMutex sync.Mutex
	eventMap      map[string]*event

	lsMapMutex sync.Mutex
	lsMap      map[string]*listeners
}

func New(conn net.Conn, config *control.Config) (e *Events) {
	ctrl := control.New(conn, config)
	e = &Events{
		Closer:   ctrl,
		ctrl:     ctrl,
		codec:    ctrl.Codec(),
		logger:   ctrl.Logger(),
		eventMap: make(map[string]*event),
		lsMap:    make(map[string]*listeners),
	}

	e.ctrl.AddFuncs(control.Funcs{
		cmdSetEvent:       e.setEvent,
		cmdTriggerEvent:   e.triggerEvent,
		cmdSetEventFilter: e.setEventFilter,
	})
	return
}

// Ready signalizes that the initialization is done.
// Events can now be triggered.
// This should be only called once.
func (e *Events) Ready() {
	e.ctrl.Ready()
}
