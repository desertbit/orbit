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

// AddSignal adds a signal with the given id to the signaler.
func (s *Signaler) AddSignal(id string) {
	s.addSignal(id)
}

// AddSignalFilter adds a signal with the given id to the signaler.
// In addition, a filter func is given that is directly set as filter
// on this signal.
func (s *Signaler) AddSignalFilter(id string, filterFunc FilterFunc) {
	signal := s.addSignal(id)
	signal.setFilterFunc(filterFunc)
}

// AddSignals adds the signals with the given ids to the signaler.
func (s *Signaler) AddSignals(ids []string) {
	s.signalsMutex.Lock()
	defer s.signalsMutex.Unlock()

	for _, id := range ids {
		// Log if the signal is overwritten.
		if _, ok := s.signals[id]; ok {
			s.logger.Printf("signal '%s' registered more than once", id)
		}
		s.signals[id] = newSignal(id)
	}
}

// TriggerSignal triggers the signal with the given id and sends the given
// data along to the peers. If data is nil, the signal carries no payload.
// Returns ErrSignalNotFound, if the signal could not be found.
func (s *Signaler) TriggerSignal(id string, data interface{}) (err error) {
	// Retrieve the signal.
	signal, err := s.getSignal(id)
	if err != nil {
		return
	}

	// In case the signal is not active (meaning no one listens for it)
	// we do not need to bother sending it.
	if !signal.isActive() {
		return
	}

	// Check if the signal is filtered out.
	var conformsToFilter bool
	conformsToFilter, err = signal.conformsToFilter(data)
	if err != nil || !conformsToFilter {
		return
	}

	// Trigger the signal with the data.
	return s.callTriggerSignal(id, data)
}

//###############//
//### Private ###//
//###############//

// getSignal returns the signal with the given id from the signaler.
// Returns ErrSignalNotFound, if the signal could not be found.
func (s *Signaler) getSignal(id string) (signal *signal, err error) {
	// Get the signal thread-safe.
	s.signalsMutex.Lock()
	signal = s.signals[id]
	s.signalsMutex.Unlock()

	// Check if the signal has been found.
	if signal == nil {
		err = ErrSignalNotFound
	}
	return
}

// addSignal adds a signal with the given id to the signaler and returns
// the signal afterwards.
func (s *Signaler) addSignal(id string) (ev *signal) {
	// Create the signal.
	ev = newSignal(id)

	s.signalsMutex.Lock()
	// Log if the signal is overwritten.
	if _, ok := s.signals[id]; ok {
		s.logger.Printf("signal '%s' registered more than once", id)
	}
	// Save the signal.
	s.signals[id] = ev
	s.signalsMutex.Unlock()

	return
}

// callTriggerSignal triggers the signal with the given id and sends along
// the given data. This function does not block.
// The data is encoded using the signaler's codec.
func (s *Signaler) callTriggerSignal(id string, data interface{}) (err error) {
	var dataBytes []byte

	// Encode the data.
	if data != nil {
		dataBytes, err = s.codec.Encode(data)
		if err != nil {
			return err
		}
	}

	// Trigger the signal in a non-blocking way.
	return s.ctrl.CallOneWay(cmdTriggerSignal, &api.TriggerSignal{
		ID:   id,
		Data: dataBytes,
	})
}

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// setSignalState is a control func that is callable from the remote peer.
// It sets the state of the signal with the sent id to either active
// or inactive. An inactive signal can not be triggered.
func (s *Signaler) setSignalState(ctx *control.Context) (interface{}, error) {
	// Decode the data.
	var data api.SetSignal
	err := ctx.Decode(&data)
	if err != nil {
		return nil, control.Err(err, "internal error", 1)
	}

	// Retrieve the signal.
	signal, err := s.getSignal(data.ID)
	if err != nil {
		return nil, control.Err(err, "signal does not exist", 2)
	}

	// Set its state to active
	signal.setActive(data.Active)
	return nil, nil
}

// setSignalFilter is a control func that as callable from the remote peer.
// It sets a filter on the signal.
func (s *Signaler) setSignalFilter(ctx *control.Context) (v interface{}, err error) {
	// Decode the data.
	var data api.SetSignalFilter
	err = ctx.Decode(&data)
	if err != nil {
		err = control.Err(err, "internal error", 1)
		return
	}

	// Retrieve the signal.
	signal, err := s.getSignal(data.ID)
	if err != nil {
		err = control.Err(err, "signal does not exist", 2)
		return
	}

	// Set the filter on the signal.
	err = signal.setFilter(newContext(data.Data, s.codec))
	if err != nil {
		if err == ErrFilterFuncUndefined {
			err = control.Err(err, "filter not set", 3)
		} else {
			err = control.Err(err, "internal error", 1)
		}
		return
	}

	return
}
