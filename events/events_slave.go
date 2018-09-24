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
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/internal/api"
)

// May return ErrFilterFuncUndefined and ErrEventNotFound.
func (e *Events) SetEventFilter(id string, data interface{}) (err error) {
	return e.callSetEventFilter(id, data)
}

func (e *Events) OnEvent(id string) *Listener {
	return e.addListener(id, defaultLsChanSize, false)
}

func (e *Events) OnEventOpts(id string, channelSize int) *Listener {
	return e.addListener(id, channelSize, false)
}

func (e *Events) OnEventFunc(id string, f func(ctx *Context)) *Listener {
	l := e.addListener(id, defaultLsChanSize, false)
	l.bindFunc(f)
	return l
}

func (e *Events) OnceEvent(id string) *Listener {
	return e.addListener(id, defaultLsChanSize, true)
}

func (e *Events) OnceEventOpts(id string, channelSize int) *Listener {
	return e.addListener(id, channelSize, true)
}

func (e *Events) OnceEventFunc(id string, f func(ctx *Context)) *Listener {
	l := e.addListener(id, defaultLsChanSize, true)
	l.bindFunc(f)
	return l
}

//###############//
//### Private ###//
//###############//

func (e *Events) addListener(eventID string, chanSize int, once bool) (l *Listener) {
	var (
		ok bool
		ls *listeners
	)

	e.lsMapMutex.Lock()
	if ls, ok = e.lsMap[eventID]; !ok {
		ls = newListeners(e, eventID)
		e.lsMap[eventID] = ls
	}
	e.lsMapMutex.Unlock()

	l = newListener(ls, chanSize, once)
	return
}

// Bind to the remote peer's event and get updates.
func (e *Events) callSetEvent(id string, active bool) (err error) {
	data := api.SetEvent{
		ID:     id,
		Active: active,
	}

	_, err = e.ctrl.Call(cmdSetEvent, &data)
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok && cErr.Code == 2 {
			err = ErrEventNotFound
		}
		return
	}
	return
}

// Called if the remote peer's event has been triggered.
func (e *Events) triggerEvent(ctx *control.Context) (v interface{}, err error) {
	var data api.TriggerEvent
	err = ctx.Decode(&data)
	if err != nil {
		return
	}

	// Build the event context.
	eventCtx := newContext(data.Data, e.codec)

	// Obtain the listeners for the given event.
	var ls *listeners
	e.lsMapMutex.Lock()
	ls = e.lsMap[data.ID]
	e.lsMapMutex.Unlock()

	// Trigger the event if defined.
	if ls != nil {
		ls.trigger(eventCtx)
	}

	return
}

// Set the filter on the remote peer's event.
func (e *Events) callSetEventFilter(id string, data interface{}) (err error) {
	dataBytes, err := e.codec.Encode(data)
	if err != nil {
		return
	}

	_, err = e.ctrl.Call(cmdSetEventFilter, &api.SetEventFilter{
		ID:   id,
		Data: dataBytes,
	})
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok {
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
