/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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
	"net"
	"sync"

	"github.com/desertbit/orbit/codec"

	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/internal/api"
)

const (
	cmdSetEvent     = "SetEvent"
	cmdTriggerEvent = "TriggerEvent"
)

type Events struct {
	closer.Closer

	ctrl *control.Control

	codec codec.Codec

	// TODO: maybe rename to events...
	eventMapMutex sync.Mutex
	eventMap      map[string]*Event

	lsMapMutex sync.Mutex
	lsMap      map[string]*listeners
}

func New(conn net.Conn, config *control.Config) (e *Events) {
	ctrl := control.New(conn, config)
	e = &Events{
		Closer:   ctrl,
		ctrl:     ctrl,
		codec:    ctrl.Config().Codec,
		eventMap: make(map[string]*Event),
		lsMap:    make(map[string]*listeners),
	}

	e.ctrl.RegisterFuncs(control.Funcs{
		cmdSetEvent:     e.setEvent,
		cmdTriggerEvent: e.triggerEvent,
	})
	e.ctrl.Ready()
	return
}

// Event returns the event for the given ID.Event
func (e *Events) Event(id string) (event *Event, err error) {
	e.eventMapMutex.Lock()
	event = e.eventMap[id]
	e.eventMapMutex.Unlock()

	if e == nil {
		err = ErrEventNotFound
	}
	return
}

func (e *Events) RegisterEvent(id string) (event *Event) {
	event = newEvent(id)

	e.eventMapMutex.Lock()
	e.eventMap[id] = event
	e.eventMapMutex.Unlock()
	return
}

func (e *Events) RegisterEvents(ids []string) {
	e.eventMapMutex.Lock()
	for _, id := range ids {
		e.eventMap[id] = newEvent(id)
	}
	e.eventMapMutex.Unlock()
}

func (e *Events) TriggerEvent(id string, data interface{}) (err error) {
	event, err := e.Event(id)
	if err != nil {
		return
	}

	if event.isActive() {
		err = e.callTriggerEvent(id, data)
		if err != nil {
			return
		}
	}

	return
}

func (e *Events) OnEvent(id string) *Listener {
	return e.addListener(id, defaultLsChanSize, false, nil)
}

func (e *Events) OnEventOpts(id string, channelSize int) *Listener {
	return e.addListener(id, channelSize, false, nil)
}

func (e *Events) OnEventFunc(id string, f func(ctx *Context)) *Listener {
	l := e.addListener(id, defaultLsChanSize, false, e.CloseChan())
	go l.listenRoutine(f)
	return l
}

func (e *Events) OnceEvent(id string) *Listener {
	return e.addListener(id, defaultLsChanSize, true, nil)
}

func (e *Events) OnceEventOpts(id string, channelSize int) *Listener {
	return e.addListener(id, channelSize, true, nil)
}

func (e *Events) OnceEventFunc(id string, f func(ctx *Context)) *Listener {
	l := e.addListener(id, defaultLsChanSize, true, e.CloseChan())
	go l.listenRoutine(f)
	return l
}

//###############//
//### Private ###//
//###############//

// TODO: remove CloseChan
func (e *Events) addListener(eventID string, chanSize int, once bool, closeChan <-chan struct{}) (l *Listener) {
	var (
		ok bool
		ls *listeners
	)

	// TODO: CloseChan
	e.lsMapMutex.Lock()
	if ls, ok = e.lsMap[eventID]; !ok {
		ls = newListeners(e.CloseChan(), e, eventID)
		e.lsMap[eventID] = ls
	}
	e.lsMapMutex.Unlock()

	// TODO: CloseChan
	// TODO: this combo shout be moved into a helper func.
	l = newListener(ls, chanSize, once, closeChan)
	ls.Add(l)

	// Ensure the event is triggered
	ls.activeChan <- true

	return
}

func (e *Events) callSetEvent(id string, active bool) (err error) {
	data := api.SetEvent{
		ID:     id,
		Active: active,
	}

	// TODO: set timeout!
	_, err = e.ctrl.Call(cmdSetEvent, &data)
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok && cErr.Code == 2 {
			err = ErrEventNotFound
		}
		return
	}
	return
}

func (e *Events) setEvent(c *control.Context) (interface{}, error) {
	var data api.SetEvent
	err := c.Decode(&data)
	if err != nil {
		return nil, control.Err(err, "internal error", 1)
	}

	event, err := e.Event(data.ID)
	if err != nil {
		return nil, control.Err(err, "event does not exists", 2)
	}

	event.setActive(data.Active)
	return nil, nil
}

func (e *Events) callTriggerEvent(id string, data interface{}) error {
	dataBytes, err := e.codec.Encode(data)
	if err != nil {
		return err
	}

	return e.ctrl.OneShot(cmdTriggerEvent, &api.TriggerEvent{
		ID:   id,
		Data: dataBytes,
	})
}

func (e *Events) triggerEvent(ctx *control.Context) (v interface{}, err error) {
	var data api.TriggerEvent
	err = ctx.Decode(&data)
	if err != nil {
		return
	}

	// Build the event context.
	eventCtx := newContext(data.Data, e.codec)

	// TODO: Lock
	// TODO: what if the event is not present? This will panic!
	// TODO: This should be expanded to several lines...
	// Now inform all listeners that are interested in this event.
	for _, listener := range e.lsMap[data.ID].lMap {
		listener.handleEvent(eventCtx)
	}

	return
}
