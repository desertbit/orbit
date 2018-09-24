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

package roe

import (
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/roc"
)

// May return ErrFilterFuncUndefined and ErrEventNotFound.
func (r *ROE) SetEventFilter(id string, data interface{}) (err error) {
	return r.callSetEventFilter(id, data)
}

func (r *ROE) OnEvent(id string) *Listener {
	return r.addListener(id, defaultLsChanSize, false)
}

func (r *ROE) OnEventOpts(id string, channelSize int) *Listener {
	return r.addListener(id, channelSize, false)
}

func (r *ROE) OnEventFunc(id string, f func(ctx *Context)) *Listener {
	l := r.addListener(id, defaultLsChanSize, false)
	l.bindFunc(f)
	return l
}

func (r *ROE) OnceEvent(id string) *Listener {
	return r.addListener(id, defaultLsChanSize, true)
}

func (r *ROE) OnceEventOpts(id string, channelSize int) *Listener {
	return r.addListener(id, channelSize, true)
}

func (r *ROE) OnceEventFunc(id string, f func(ctx *Context)) *Listener {
	l := r.addListener(id, defaultLsChanSize, true)
	l.bindFunc(f)
	return l
}

//###############//
//### Private ###//
//###############//

func (r *ROE) addListener(eventID string, chanSize int, once bool) (l *Listener) {
	var (
		ok bool
		ls *listeners
	)

	r.lsMapMutex.Lock()
	if ls, ok = r.lsMap[eventID]; !ok {
		ls = newListeners(r, eventID)
		r.lsMap[eventID] = ls
	}
	r.lsMapMutex.Unlock()

	l = newListener(ls, chanSize, once)
	return
}

// Bind to the remote peer's event and get updates.
func (r *ROE) callSetEvent(id string, active bool) (err error) {
	data := api.SetEvent{
		ID:     id,
		Active: active,
	}

	_, err = r.ctrl.Call(cmdSetEvent, &data)
	if err != nil {
		if cErr, ok := err.(*roc.ErrorCode); ok && cErr.Code == 2 {
			err = ErrEventNotFound
		}
		return
	}
	return
}

// Set the filter on the remote peer's event.
func (r *ROE) callSetEventFilter(id string, data interface{}) (err error) {
	dataBytes, err := r.codec.Encode(data)
	if err != nil {
		return
	}

	_, err = r.ctrl.Call(cmdSetEventFilter, &api.SetEventFilter{
		ID:   id,
		Data: dataBytes,
	})
	if err != nil {
		if cErr, ok := err.(*roc.ErrorCode); ok {
			if cErr.Code == 2 {
				err = ErrEventNotFound
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

// Called if the remote peer's event has been triggered.
func (r *ROE) triggerEvent(ctx *roc.Context) (v interface{}, err error) {
	var data api.TriggerEvent
	err = ctx.Decode(&data)
	if err != nil {
		return
	}

	// Build the event context.
	eventCtx := newContext(data.Data, r.codec)

	// Obtain the listeners for the given event.
	var ls *listeners
	r.lsMapMutex.Lock()
	ls = r.lsMap[data.ID]
	r.lsMapMutex.Unlock()

	// Trigger the event if defined.
	if ls != nil {
		ls.trigger(eventCtx)
	}

	return
}
