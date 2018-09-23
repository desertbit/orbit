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
)

type Event struct {
	id        string
	active bool // bind state of the peer.
	mutex sync.Mutex
}

func newEvent(id string) *Event {
	return &Event{
		id: id,
	}
}

func (e *Event) setActive(active bool) {
	e.mutex.Lock()
	e.active = active
	e.mutex.Unlock()
}

func (e *Event) isActive() (active bool) {
	e.mutex.Lock()
	active = e.active
	e.mutex.Unlock()
	return
}
