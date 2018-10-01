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

// The Filter type is a function that takes the trigger data
// and returns whether the data conforms to some predefined
// conditions, and is allowed to be triggered.
// This func is returned by the FilterFunc to create a
// static filter that is fed with dynamic data at runtime.
type Filter func(data interface{}) (conforms bool, err error)

// The FilterFunc type takes a request context and returns
// a new filter func that will be set as the signal's filter.
type FilterFunc func(ctx *Context) (f Filter, err error)

// The signal type represents one signal that can be triggered.
// It allows to define a filter on it, to control when the
// signal can actually be triggered.
type signal struct {
	// The id of the signal.
	id string

	// Synchronises the access to the active flag, the filter
	// and filter func.
	mutex sync.Mutex
	// Whether this signal is currently active and can be triggered.
	active bool
	// The filter func that produces the filter for the signal.
	filterFunc FilterFunc
	// The actual filter of the signal that is fed with the
	// data of the trigger request and determines, if the
	// signal is triggered or not.
	filter Filter
}

// newSignal returns a new signal with the given id.
func newSignal(id string) *signal {
	return &signal{
		id: id,
	}
}

// setActive sets the active state of the signal.
// This function is thread-safe.
func (s *signal) setActive(active bool) {
	s.mutex.Lock()
	s.active = active
	s.mutex.Unlock()
}

// isActive returns the current active state of the signal.
// This function is thread-safe.
func (s *signal) isActive() (active bool) {
	s.mutex.Lock()
	active = s.active
	s.mutex.Unlock()
	return
}

// setFilterFunc sets the filter func of the signal.
// This function is thread-safe.
func (s *signal) setFilterFunc(filterFunc FilterFunc) {
	s.mutex.Lock()
	s.filterFunc = filterFunc
	s.mutex.Unlock()
}

// setFilter calls the filter func of this signal with the given
// request context and produces a new filter that is saved
// in the signal.
// Returns ErrFilterFuncUndefined, if no filter func is defined.
// This function is thread-safe.
func (s *signal) setFilter(ctx *Context) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if a filter func is defined.
	if s.filterFunc == nil {
		err = ErrFilterFuncUndefined
	}

	// Produce and save a new filter.
	s.filter, err = s.filterFunc(ctx)
	return
}

// conformsToFilter is a convenience function that takes the data
// of a trigger request and returns, whether the data conforms
// to the filter defined on the signal.
func (s *signal) conformsToFilter(data interface{}) (conforms bool, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If no filter is defined, the signal conforms by default.
	if s.filter == nil {
		return true, nil
	}

	// Returns the output of the filter for the trigger data.
	return s.filter(data)
}
