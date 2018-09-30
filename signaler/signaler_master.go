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
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/internal/api"
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
	if signal.isActive() {
		// Check if the signal is filtered out.
		var conformsToFilter bool
		conformsToFilter, err = signal.conformsToFilter(data)
		if err != nil || !conformsToFilter {
			return
		}

		// Trigger the signal with the data.
		err = s.callTriggerSignal(id, data)
		if err != nil {
			return
		}
	}

	return
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
	if s == nil {
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
func (s *Signaler) callTriggerSignal(id string, data interface{}) error {
	// Encode the data.
	dataBytes, err := s.codec.Encode(data)
	if err != nil {
		return err
	}

	// Trigger the signal in a non-blocking way.
	return s.ctrl.CallAsync(cmdTriggerSignal, &api.TriggerSignal{
		ID:   id,
		Data: dataBytes,
	})
}

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// setSignal is a control func that is callable from the remote peer.
// It signals that the remote peer is either interested or not (anymore)
// interested in the signal, based on the active flag in the request data.
// It ensures that the signal's state reflects the wish of the remote peer,
// meaning it is set to active, if the remote peer is interested, or, set
// to inactive when the remote peer is not interested and if he was the last
// one to be so.
func (s *Signaler) setSignal(ctx *control.Context) (interface{}, error) {
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
