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

package signaler

import (
	"sync"
)

type Filter func(data interface{}) (conforms bool, err error)

type FilterFunc func(ctx *Context) (f Filter, err error)

type signal struct {
	id string

	mutex      sync.Mutex
	active     bool // bind state of the peer.
	filterFunc FilterFunc
	filter     Filter
}

func newSignal(id string) *signal {
	return &signal{
		id: id,
	}
}

func (s *signal) setActive(active bool) {
	s.mutex.Lock()
	s.active = active
	s.mutex.Unlock()
}

func (s *signal) isActive() (active bool) {
	s.mutex.Lock()
	active = s.active
	s.mutex.Unlock()
	return
}

func (s *signal) setFilterFunc(filterFunc FilterFunc) {
	s.mutex.Lock()
	s.filterFunc = filterFunc
	s.mutex.Unlock()
}

func (s *signal) setFilter(ctx *Context) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.filterFunc == nil {
		err = ErrFilterFuncUndefined
		return
	}

	s.filter, err = s.filterFunc(ctx)
	return
}

func (s *signal) conformsToFilter(data interface{}) (conforms bool, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.filter == nil {
		return true, nil
	}

	return s.filter(data)
}
