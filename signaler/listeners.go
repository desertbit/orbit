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
package signaler

import (
	"sync"
)

// The listeners type represents several single listeners that are
// listening on the same signal. It manages the state of the signal,
// meaning that it sets it inactive, if all listeners are gone, and
// switches it to active, if at least one listener returns.
type listeners struct {
	// The signaler that manages the signals.
	s *Signaler
	// The id of the signal.
	signalID string

	// Synchronises the access to the listeners map and the
	// id counter.
	lMapMutex sync.Mutex
	// The listener types that are handled by this listeners.
	// The key of the map is the id of the listener.
	lMap map[uint64]*Listener
	// A counter that is used to produce new ids for new
	// listeners.
	idCount uint64

	// This channel is used to manage the active state of the
	// signal. Whenever a listener joins or leaves the listeners,
	// over this channel a value is sent that triggers a check to
	// whether the signal must be set to in-/active.
	// Buffered channel.
	activeChan chan struct{}
	// This channel is used to remove a listener from the listeners.
	// Buffered channel.
	removeChan chan uint64
	// This channel is the close channel of the signaler, which
	// can trigger the shutdown of the listeners.
	closeChan <-chan struct{}
}

// newListeners returns a new listeners for the given signalID and
// with a reference to the signaler.
func newListeners(e *Signaler, signalID string) *listeners {
	ls := &listeners{
		s:          e,
		signalID:   signalID,
		lMap:       make(map[uint64]*Listener),
		activeChan: make(chan struct{}, 1),
		removeChan: make(chan uint64, 3),
		closeChan:  e.CloseChan(),
	}

	// Start the main routine that takes care of handling the active
	// state changes of the signal and the removal of listeners.
	go ls.routine()

	return ls
}

// add adds the listener to our listeners type. In case the
// signal is currently inactive, the signal is switched back
// to active.
func (ls *listeners) add(l *Listener) {
	// Create a new listener ID and ensure it is unqiue.
	// Add it to the listeners map and set the ID.
	//
	// WARNING: Possible loop, if more than 2^64 listeners
	// have been registered. Refactor in 25 years.
	ls.lMapMutex.Lock()

	for {
		l.id = ls.idCount
		ls.idCount++

		if _, ok := ls.lMap[l.id]; !ok {
			break
		}
	}

	// Add the listener to our map
	ls.lMap[l.id] = l

	// Activate the signal.
	ls.activateIfRequired()

	ls.lMapMutex.Unlock()
}

// trigger triggers all listeners that are currently part of this
// listeners struct with the given request context.
func (ls *listeners) trigger(ctx *Context) {
	ls.lMapMutex.Lock()
	for _, l := range ls.lMap {
		l.trigger(ctx)
	}
	ls.lMapMutex.Unlock()
}

// activateIfRequired simply writes a value into the active channel of
// this listeners in a non-blocking way.
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
			// This default case is crucial, to avoid blocking when
			// another routine has written a value into the buffer before
			// we could write our own.
		}
	}
}

// routine is the main worker routine of this listeners and should be started
// exactly once, ideally when creating a new listeners.
// It watches the three channels (close, active, remove) for new values and
// performs the necessary actions for each of them.
//
// If the close channel has a value, it shuts down the listeners.
//
// If the active channel has a value, it checks whether the active state
// of the signal has to change. This is the case, if the signal is currently
// inactive and a new listener has been added, or the signal is currently
// active, but the last listener has left. In case the active state must
// change, it calls the callSetSignalState function to change the state.
//
// If the remove channel has a value, it removes the listener from its map
// with the id sent over the channel.
func (ls *listeners) routine() {
	// Set all listener to off on exit.
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
			// We must check whether the active state of the signal must be changed.
			ls.lMapMutex.Lock()
			activate = len(ls.lMap) != 0
			ls.lMapMutex.Unlock()

			// Only proceed, if the state changed.
			if isActive == activate {
				continue Loop
			}
			isActive = activate

			// Adjust the state of the signal to the new state.
			err = ls.s.callSetSignalState(ls.signalID, isActive)
			if err != nil {
				ls.s.logger.Printf("listeners signal '%s': callSetSignalState error: %v", ls.signalID, err)
				return
			}

		case id := <-ls.removeChan:
			// Remove the listener with the received id from our map.
			ls.lMapMutex.Lock()
			delete(ls.lMap, id)

			// Deactivate the signal if no listeners are left.
			if len(ls.lMap) == 0 {
				ls.activateIfRequired()
			}
			ls.lMapMutex.Unlock()
		}
	}
}
