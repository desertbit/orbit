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

type Filter func(data interface{}) (conforms bool, err error)

type FilterFunc func(ctx *Context) (f Filter, err error)

type Event struct {
	id string

	activeMutex sync.Mutex
	active      bool // bind state of the peer.

	filterMutex sync.Mutex
	filterFunc  FilterFunc
	filter      Filter
}

func newEvent(id string) *Event {
	return &Event{
		id: id,
	}
}

func (e *Event) setActive(active bool) {
	e.activeMutex.Lock()
	e.active = active
	e.activeMutex.Unlock()
}

func (e *Event) isActive() (active bool) {
	e.activeMutex.Lock()
	active = e.active
	e.activeMutex.Unlock()
	return
}

func (e *Event) setFilterFunc(filterFunc FilterFunc) {
	e.filterMutex.Lock()
	e.filterFunc = filterFunc
	e.filterMutex.Unlock()
}

func (e *Event) setFilter(ctx *Context) (err error) {
	e.filterMutex.Lock()
	defer e.filterMutex.Unlock()

	if e.filterFunc == nil {
		err = errFilterFuncUndefined
		return
	}

	e.filter, err = e.filterFunc(ctx)
	if err != nil {
		return
	}

	return
}

func (e *Event) conformsToFilter(data interface{}) (conforms bool, err error) {
	e.filterMutex.Lock()
	conforms, err = e.filter(data)
	if err != nil {
		return
	}
	e.filterMutex.Unlock()

	return
}
