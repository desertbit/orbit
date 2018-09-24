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

func (s *Signaler) AddSignal(id string) {
	s.addSignal(id)
}

func (s *Signaler) AddSignalFilter(id string, filterFunc FilterFunc) {
	signal := s.addSignal(id)
	signal.setFilterFunc(filterFunc)
}

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

func (s *Signaler) TriggerSignal(id string, data interface{}) (err error) {
	signal, err := s.getSignal(id)
	if err != nil {
		return
	}

	if signal.isActive() {
		// Check if the signal is filtered out
		var conformsToFilter bool
		conformsToFilter, err = signal.conformsToFilter(data)
		if err != nil || !conformsToFilter {
			return
		}

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

func (s *Signaler) getSignal(id string) (signal *signal, err error) {
	s.signalsMutex.Lock()
	signal = s.signals[id]
	s.signalsMutex.Unlock()

	if s == nil {
		err = ErrSignalNotFound
	}
	return
}

func (s *Signaler) addSignal(id string) (ev *signal) {
	ev = newSignal(id)

	s.signalsMutex.Lock()
	defer s.signalsMutex.Unlock()

	// Log if the signal is overwritten.
	if _, ok := s.signals[id]; ok {
		s.logger.Printf("signal '%s' registered more than once", id)
	}

	s.signals[id] = ev
	return
}

// Call the listeners on the remote peer.
func (s *Signaler) callTriggerSignal(id string, data interface{}) error {
	dataBytes, err := s.codec.Encode(data)
	if err != nil {
		return err
	}

	return s.ctrl.OneShot(cmdTriggerSignal, &api.TriggerSignal{
		ID:   id,
		Data: dataBytes,
	})
}

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// Called if the remote peer wants to be informed about the given signal.
func (s *Signaler) setSignal(c *control.Context) (interface{}, error) {
	var data api.SetSignal
	err := c.Decode(&data)
	if err != nil {
		return nil, control.Err(err, "internal error", 1)
	}

	signal, err := s.getSignal(data.ID)
	if err != nil {
		return nil, control.Err(err, "signal does not exists", 2)
	}

	signal.setActive(data.Active)
	return nil, nil
}

// Called when the remote peer wants to set a filter on an signal.
func (s *Signaler) setSignalFilter(ctx *control.Context) (v interface{}, err error) {
	var data api.SetSignalFilter
	err = ctx.Decode(&data)
	if err != nil {
		err = control.Err(err, "internal error", 1)
		return
	}

	signal, err := s.getSignal(data.ID)
	if err != nil {
		err = control.Err(err, "signal does not exists", 2)
		return
	}

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
