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
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/control"
)

// SetSignalFilter sets the filter data on the signal with the given id.
// The data will be passed to the filter function and determine, when
// the signal is triggered.
// Returns ErrFilterFuncUndefined, if no filter func is defined on the signal.
// Returns ErrSignalNotFound, if the signal does not exist.
func (s *Signaler) SetSignalFilter(id string, data interface{}) (err error) {
	return s.callSetSignalFilter(id, data)
}

// OnSignal adds a listener to the signal with the given id.
// The Listener will output a request context on its channel
// when the signal has been triggered.
func (s *Signaler) OnSignal(id string) *Listener {
	return s.addListener(id, defaultLsChanSize, false)
}

// OnSignalOpts adds a listener like OnSignal() does, but
// allows to configure the channel size of the listener. This
// determines how many events can be buffered in the listener before
// the triggerSignal handler func will block and further events can not
// be processed.
// The size of the channel must be greater than 0 (Unbuffered channels
// are not allowed).
func (s *Signaler) OnSignalOpts(id string, channelSize int) *Listener {
	return s.addListener(id, channelSize, false)
}

// OnSignalFunc adds a listener like OnSignal() does, but takes
// a function that is executed whenever the signal is triggered.
func (s *Signaler) OnSignalFunc(id string, f func(ctx *Context)) *Listener {
	l := s.addListener(id, defaultLsChanSize, false)
	l.bindFunc(f)
	return l
}

// OnceSignal adds a listener like OnSignal() does, but the listener
// can only be triggered once and gets removed after the first event.
func (s *Signaler) OnceSignal(id string) *Listener {
	return s.addListener(id, defaultLsChanSize, true)
}

// OnceSignalOpts adds a listener like OnceSignal() does, but allows
// to configure the channel size of the listener. This
// determines how many events can be buffered in the listener before
// the triggerSignal handler func will block and further events can not
// be processed.
// The size of the channel must be greater than 0 (Unbuffered channels
// are not allowed).
func (s *Signaler) OnceSignalOpts(id string, channelSize int) *Listener {
	return s.addListener(id, channelSize, true)
}

// OnSignalFunc adds a listener like OnceSignal() does, but takes
// a function that is executed when the signal is triggered.
func (s *Signaler) OnceSignalFunc(id string, f func(ctx *Context)) *Listener {
	l := s.addListener(id, defaultLsChanSize, true)
	l.bindFunc(f)
	return l
}

//###############//
//### Private ###//
//###############//

// addListener adds a listener to the signal with the given id.
// The listener will have an event channel with the given size.
// If once is true, the listener will only be notified of a signal
// trigger once.
func (s *Signaler) addListener(id string, chanSize int, once bool) (l *Listener) {
	var (
		ok bool
		ls *listeners
	)

	// Check if a listeners has been created for this signal already.
	// If not, create a new one.
	s.lsMapMutex.Lock()
	if ls, ok = s.lsMap[id]; !ok {
		ls = newListeners(s, id)
		s.lsMap[id] = ls
	}
	s.lsMapMutex.Unlock()

	// Create a new listener.
	l = newListener(ls, chanSize, once)
	return
}

// callSetSignalState calls the setSignalState control func on the remote peer
// for the signal with the given id and sets the state of the signal
// to either active or inactive. An inactive signal can not be triggered.
// Returns ErrSignalNotFound, if the signal does not exist.
func (s *Signaler) callSetSignalState(id string, active bool) (err error) {
	// Create the data.
	data := api.SetSignal{
		ID:     id,
		Active: active,
	}

	// Call the control func to set the signal's state.
	_, err = s.ctrl.Call(cmdSetSignalState, &data)
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok && cErr.Code == 2 {
			err = ErrSignalNotFound
		}
		return
	}
	return
}

// callSetSignalFilter calls the setSignalFilter control func on the remote
// peer for the signal with the given id and calls the filter function
// with the provided data.
// Returns ErrFilterFuncUndefined, if no filter func is defined on the signal.
// Returns ErrSignalNotFound, if the signal does not exist.
func (s *Signaler) callSetSignalFilter(id string, data interface{}) (err error) {
	// Encode the data.
	dataBytes, err := s.codec.Encode(data)
	if err != nil {
		return
	}

	// Call the control func to set the signal's filter
	_, err = s.ctrl.Call(cmdSetSignalFilter, &api.SetSignalFilter{
		ID:   id,
		Data: dataBytes,
	})
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok {
			if cErr.Code == 2 {
				err = ErrSignalNotFound
			} else if cErr.Code == 3 {
				err = ErrFilterFuncUndefined
			}
		}
		return
	}

	return
}

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// triggerSignal is a control func that can be called by the remote peer,
// when a signal has been triggered. All listeners that previously stated
// their interest in the signal will be notified and get passed the data
// sent in the trigger request.
func (s *Signaler) triggerSignal(ctx *control.Context) (v interface{}, err error) {
	// Decode the data.
	var data api.TriggerSignal
	err = ctx.Decode(&data)
	if err != nil {
		return
	}

	// Build the signal context.
	signalCtx := newContext(data.Data, s.codec)

	// Obtain the listeners for the given signal.
	var ls *listeners
	s.lsMapMutex.Lock()
	ls = s.lsMap[data.ID]
	s.lsMapMutex.Unlock()

	// Trigger the signal if defined.
	if ls != nil {
		ls.trigger(signalCtx)
	}

	return
}
