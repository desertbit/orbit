/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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
		return
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
