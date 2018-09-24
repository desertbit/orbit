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
	"sync"
)

type Filter func(data interface{}) (conforms bool, err error)

type FilterFunc func(ctx *Context) (f Filter, err error)

type event struct {
	id string

	mutex      sync.Mutex
	active     bool // bind state of the peer.
	filterFunc FilterFunc
	filter     Filter
}

func newEvent(id string) *event {
	return &event{
		id: id,
	}
}

func (e *event) setActive(active bool) {
	e.mutex.Lock()
	e.active = active
	e.mutex.Unlock()
}

func (e *event) isActive() (active bool) {
	e.mutex.Lock()
	active = e.active
	e.mutex.Unlock()
	return
}

func (e *event) setFilterFunc(filterFunc FilterFunc) {
	e.mutex.Lock()
	e.filterFunc = filterFunc
	e.mutex.Unlock()
}

func (e *event) setFilter(ctx *Context) (err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.filterFunc == nil {
		err = ErrFilterFuncUndefined
		return
	}

	e.filter, err = e.filterFunc(ctx)
	return
}

func (e *event) conformsToFilter(data interface{}) (conforms bool, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.filter == nil {
		return true, nil
	}

	return e.filter(data)
}
