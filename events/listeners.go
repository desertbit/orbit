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
package events

import (
	"sync"
	"sync/atomic"
)

type listeners struct {
	e       *Events
	eventID string

	idCount   uint64
	lMap      map[uint64]*Listener
	lMapMutex sync.Mutex // TODO:

	activeChan chan bool
}

func newListeners(closeChan <-chan struct{}, e *Events, eventID string) *listeners {
	ls := &listeners{
		e:          e,
		eventID:    eventID,
		lMap:       make(map[uint64]*Listener),
		activeChan: make(chan bool, 1),
	}

	// Start the active routine that takes care of switching
	go ls.activeRoutine(closeChan)

	return ls
}

func (ls *listeners) Add(l *Listener) {
	// Create a new listener ID and ensure it is unqiue.
	// Add it to the listeners map and set the ID.
	//
	// WARNING: Possible loop, if more than 2^64 listeners
	// have been registered. Refactor in 25 years.
	var id uint64

	// TODO: Lock

	for {
		if _, ok := ls.lMap[id]; !ok {
			break
		}

		id = atomic.AddUint64(&ls.idCount, 1) // TODO: remove
	}

	l.id = id
	ls.lMap[id] = l
}

func (ls *listeners) Remove(id uint64) {
	ls.lMapMutex.Lock()

	delete(ls.lMap, id)

	// TODO: move to worker
	// Deactivate the event if no listeners are left
	if len(ls.lMap) == 0 {
		ls.activeChan <- false
	}

	ls.lMapMutex.Unlock()
}

// TODO: release?
func (ls *listeners) activeRoutine(closeChan <-chan struct{}) {
	var (
		err    error
		active bool
	)

	for {
		select {
		case <-closeChan:
			return

		case active = <-ls.activeChan:
			err = ls.e.callSetEvent(ls.eventID, active)
			if err != nil {
				return
			}
		}
	}
}
