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

// The Group type represents a group of signalers.
// It is useful for triggering the same signal with some data
// on all signalers contained in the group.
type Group struct {
	// Synchronises the access to the signalers map and the idCount.
	mutex sync.RWMutex
	// The signalers that are part of this group. The key to the map
	// is an id generated from the idCount property.
	signalers map[uint64]*Signaler
	// A counter that is used to produce new ids for new signalers.
	idCount uint64
}

// NewGroup creates an empty group.
func NewGroup() *Group {
	return &Group{
		signalers: make(map[uint64]*Signaler, 0),
	}
}

// Add adds the given signalers to the group.
// When a signaler is closed, it is automatically
// removed from the group. There is no need to call Remove()
// in that case.
func (g *Group) Add(ss ...*Signaler) {
	if len(ss) == 0 {
		return
	}

	g.mutex.Lock()
	for _, s := range ss {
		// Ignore nil signalers.
		if s == nil {
			continue
		}

		// Add the signaler to our group with the current
		// value of idCount as id.
		s.groupID = g.idCount
		g.signalers[s.groupID] = s
		// Generate the next id for the next signaler.
		g.idCount++

		// Remove the signaler from the group once it closes.
		s.OnClose(func() error {
			g.Remove(s)
			return nil
		})
	}
	g.mutex.Unlock()
}

// Remove removes the signaler from this group.
// If the signaler is not in the group, this is a no-op.
func (g *Group) Remove(s *Signaler) {
	g.mutex.Lock()
	delete(g.signalers, s.groupID)
	g.mutex.Unlock()
}

// Trigger calls each signaler's TriggerSignal() method of this group
// with the given id and the given data.
// If a signaler in the group does not contain the signal with the given id,
// the error is ignored and all other signalers are still triggered.
//
// Signalers can be explicitly excluded from being triggered.
func (g *Group) Trigger(id string, data interface{}, exclude ...*Signaler) (err error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	// Trigger the signal for all signalers.
TriggerLoop:
	for _, s := range g.signalers {
		// Do not trigger the signal for the excluded signalers.
		for _, ex := range exclude {
			if ex.groupID == s.groupID {
				continue TriggerLoop
			}
		}

		// It is ignored, if one of the signalers does not contain the desired signal.
		err = s.TriggerSignal(id, data)
		if err != nil {
			if err == ErrSignalNotFound {
				err = nil
				continue TriggerLoop
			}
			return
		}
	}

	return
}
