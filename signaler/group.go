/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package signaler

import (
	"sync"
)

// The Group type TODO
type Group struct {
	// TODO closer? stricty not needed, since signalers are closed, when underlying control is closed, which closes, when session closes anyways, but we could go safe.

	mutex     sync.RWMutex // TODO: Normal Mutex?
	signalers map[uint64]*Signaler
	// A counter that is used to produce new ids for new signalers.
	idCount uint64
}

// NewGroup TODO
func NewGroup() *Group {
	return &Group{
		signalers: make(map[uint64]*Signaler, 0),
	}
}

// Add TODO
// When the signaler closes, it is automatically removed from the group.
func (g *Group) Add(s *Signaler) {
	g.mutex.Lock()

	s.groupID = g.idCount
	g.signalers[s.groupID] = s
	g.idCount++

	g.mutex.Unlock()

	// Remove the signaler from the group once it closes.
	s.OnClose(func() error {
		g.Remove(s)
		return nil
	})

	return
}

// Remove TODO
// If the id is unknown, this is a no-op.
func (g *Group) Remove(s *Signaler) {
	g.mutex.Lock()
	delete(g.signalers, s.groupID)
	/*for i := range g.signalers { TODO only needed if slice used.
		if g.signalers[i] == s {
			// Remove the element from the slice at the index.
			// We are using the safe way documented in the golang slice
			// tricks (https://github.com/golang/go/wiki/SliceTricks),
			// since we must avoid a memory leak due to pointers in a slice.
			copy(g.signalers[i:], g.signalers[i+1:])
			g.signalers[len(g.signalers)-1] = nil
			g.signalers = g.signalers[:len(g.signalers)-1]
			break
		}
	}*/
	g.mutex.Unlock()
}

// Trigger TODO
// TODO exclude is not the best solution
func (g *Group) Trigger(id string, data interface{}, exclude ...*Signaler) (err error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

TriggerLoop:
	for _, s := range g.signalers {
		// Do not trigger the signal for the excluded signalers.
		for _, ex := range exclude {
			if ex.groupID == s.groupID {
				continue TriggerLoop
			}
		}

		// Trigger the signal with id at all signalers with the given data.
		// It is ignored, if one of the signalers does not contain the desired signal.
		err = s.TriggerSignal(id, data)
		if err != nil && err != ErrSignalNotFound {
			return
		}
	}

	return
}
