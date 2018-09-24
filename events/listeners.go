/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers@desertbit.com>
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
)

type listeners struct {
	e       *Events
	eventID string

	lMapMutex sync.Mutex
	lMap      map[uint64]*Listener
	idCount   uint64

	activeChan chan struct{}
	removeChan chan uint64
	closeChan  <-chan struct{}
}

func newListeners(e *Events, eventID string) *listeners {
	ls := &listeners{
		e:          e,
		eventID:    eventID,
		lMap:       make(map[uint64]*Listener),
		activeChan: make(chan struct{}, 1),
		removeChan: make(chan uint64, 3),
		closeChan:  e.CloseChan(),
	}

	// Start the active routine that takes care of switching
	go ls.activeRoutine()

	return ls
}

func (ls *listeners) add(l *Listener) {
	// Create a new listener ID and ensure it is unqiue.
	// Add it to the listeners map and set the ID.
	//
	// WARNING: Possible loop, if more than 2^64 listeners
	// have been registered. Refactor in 25 years.
	ls.lMapMutex.Lock()
	defer ls.lMapMutex.Unlock()

	for {
		l.id = ls.idCount
		ls.idCount++

		if _, ok := ls.lMap[l.id]; !ok {
			break
		}
	}

	ls.lMap[l.id] = l

	// Activate the event.
	ls.activateIfRequired()
}

func (ls *listeners) trigger(ctx *Context) {
	ls.lMapMutex.Lock()
	for _, l := range ls.lMap {
		l.trigger(ctx)
	}
	ls.lMapMutex.Unlock()
}

func (ls *listeners) activateIfRequired() {
	// Remove the oldest value if full.
	// Never block!
	select {
	case ls.activeChan <- struct{}{}:
	default:
		select {
		case <-ls.activeChan:
		default:
		}

		select {
		case ls.activeChan <- struct{}{}:
		default:
		}
	}
}

func (ls *listeners) activeRoutine() {
	// Set all events to off on exit.
	defer func() {
		ls.lMapMutex.Lock()
		for _, l := range ls.lMap {
			l.Off()
		}
		ls.lMapMutex.Unlock()
	}()

	var (
		err      error
		activate bool
		isActive bool
	)

Loop:
	for {
		select {
		case <-ls.closeChan:
			return

		case <-ls.activeChan:
			ls.lMapMutex.Lock()
			activate = len(ls.lMap) != 0
			ls.lMapMutex.Unlock()

			// Only proceed, if the state changed.
			if isActive == activate {
				continue Loop
			}
			isActive = activate

			err = ls.e.callSetEvent(ls.eventID, isActive)
			if err != nil {
				ls.e.logger.Printf("listeners event '%s': callSetEvent error: %v", ls.eventID, err)
				return
			}

		case id := <-ls.removeChan:
			ls.lMapMutex.Lock()
			delete(ls.lMap, id)
			if len(ls.lMap) == 0 {
				ls.activateIfRequired() // Deactivate the event if no listeners are left
			}
			ls.lMapMutex.Unlock()
		}
	}
}
