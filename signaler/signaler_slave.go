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

// May return ErrFilterFuncUndefined and ErrSignalNotFound.
func (s *Signaler) SetSignalFilter(id string, data interface{}) (err error) {
	return s.callSetSignalFilter(id, data)
}

func (s *Signaler) OnSignal(id string) *Listener {
	return s.addListener(id, defaultLsChanSize, false)
}

func (s *Signaler) OnSignalOpts(id string, channelSize int) *Listener {
	return s.addListener(id, channelSize, false)
}

func (s *Signaler) OnSignalFunc(id string, f func(ctx *Context)) *Listener {
	l := s.addListener(id, defaultLsChanSize, false)
	l.bindFunc(f)
	return l
}

func (s *Signaler) OnceSignal(id string) *Listener {
	return s.addListener(id, defaultLsChanSize, true)
}

func (s *Signaler) OnceSignalOpts(id string, channelSize int) *Listener {
	return s.addListener(id, channelSize, true)
}

func (s *Signaler) OnceSignalFunc(id string, f func(ctx *Context)) *Listener {
	l := s.addListener(id, defaultLsChanSize, true)
	l.bindFunc(f)
	return l
}

//###############//
//### Private ###//
//###############//

func (s *Signaler) addListener(signalID string, chanSize int, once bool) (l *Listener) {
	var (
		ok bool
		ls *listeners
	)

	s.lsMapMutex.Lock()
	if ls, ok = s.lsMap[signalID]; !ok {
		ls = newListeners(s, signalID)
		s.lsMap[signalID] = ls
	}
	s.lsMapMutex.Unlock()

	l = newListener(ls, chanSize, once)
	return
}

// Bind to the remote peer's signal and get updates.
func (s *Signaler) callSetSignal(id string, active bool) (err error) {
	data := api.SetSignal{
		ID:     id,
		Active: active,
	}

	_, err = s.ctrl.Call(cmdSetSignal, &data)
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok && cErr.Code == 2 {
			err = ErrSignalNotFound
		}
		return
	}
	return
}

// Set the filter on the remote peer's signal.
func (s *Signaler) callSetSignalFilter(id string, data interface{}) (err error) {
	dataBytes, err := s.codec.Encode(data)
	if err != nil {
		return
	}

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

// Called if the remote peer's signal has been triggered.
func (s *Signaler) triggerSignal(ctx *control.Context) (v interface{}, err error) {
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
