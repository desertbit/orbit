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

package roe

import (
	"log"
	"net"
	"sync"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/roc"

	"github.com/desertbit/closer"
)

const (
	cmdSetEvent       = "SetEvent"
	cmdTriggerEvent   = "TriggerEvent"
	cmdSetEventFilter = "SetEventFilter"
)

type ROE struct {
	closer.Closer

	ctrl   *roc.ROC
	codec  codec.Codec
	logger *log.Logger

	eventsMutex sync.Mutex
	events      map[string]*event

	lsMapMutex sync.Mutex
	lsMap      map[string]*listeners
}

func New(conn net.Conn, config *Config) (r *ROE) {
	config = prepareConfig(config)

	ctrl := roc.New(conn, config.roc)
	r = &ROE{
		Closer: ctrl,
		ctrl:   ctrl,
		codec:  ctrl.Codec(),
		logger: ctrl.Logger(),
		events: make(map[string]*event),
		lsMap:  make(map[string]*listeners),
	}

	r.ctrl.AddFuncs(roc.Funcs{
		cmdSetEvent:       r.setEvent,
		cmdTriggerEvent:   r.triggerEvent,
		cmdSetEventFilter: r.setEventFilter,
	})
	return
}

// Ready signalizes that the initialization is done.
// ROEs can now be triggered.
// This should be only called once.
func (r *ROE) Ready() {
	r.ctrl.Ready()
}
