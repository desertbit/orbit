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

	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/internal/api"
)

type Events struct {
	closer.Closer

	ctrl *control.Control

	eventMapMutex sync.Mutex
	eventMap      map[string]*Event
}

func New(conn net.Conn, config *control.Config) (e *Events) {
	e = &Event{
		Closer: closer.New(),
		ctrl:   control.New(conn, config),
		m:      make(map[string]*Event),
	}
	e.OnClose(conn.Close)
	e.OnClose(e.ctrl.Close)

	e.ctrl.RegisterFuncs(control.Funcs{
		"setEvent": e.setEvent,
	})
	e.ctrl.Ready()
	return
}

func (e *Events) RegisterEvent(id string) (event *Event) {
	event = newEvent(id)

	c.eventMapMutex.Lock()
	c.eventMap[id] = event
	c.eventMapMutex.Unlock()
	return
}

func (e *Events) RegisterEvents() {
	// TODO:
}

// Returns ErrEventNotFound if the event does not exists.
func (e *Events) TriggerEvent(id string, data interface{}) (err error) {
	event, err := e.getEvent(id)
	if err != nil {
		return
	}

	switch event.getBindState() {
	case api.StateOn:

	}

	return
}

// Returns ErrEventNotFound if the event does not exists on the peer's side.
func (e *Events) OnEvent(id string) (l Listener, err error) {
	err = e.callSetEvent(id, api.StateOn)
	// TODO:
	return
}

// Returns ErrEventNotFound if the event does not exists on the peer's side.
func (e *Events) OnceEvent(id string) (l Listener, err error) {
	err = e.callSetEvent(id, api.StateOnce)
	return
}

//###############//
//### Private ###//
//###############//

func (e *Events) getEvent(id string) (e *Event, err error) {
	c.eventMapMutex.Lock()
	e = c.eventMap[id]
	c.eventMapMutex.Unlock()

	if e == nil {
		err = ErrEventNotFound
	}
	return
}

func (e *Events) callSetEvent(id string, state api.BindState) (err error) {
	data := api.SetEvent{
		ID:    id,
		State: state,
	}

	// TODO: set timeout!
	_, err = e.ctrl.Call("setEvent", &data)
	if err != nil {
		if cErr, ok := err.(*control.ErrorCode); ok && cErr.Code() == 2 {
			err = ErrEventNotFound
		}
		return
	}
	return
}

// TODO: control rename to cmd

func (e *Events) setEvent(c *control.Context) (interface{}, error) {
	var data api.SetEvent
	err := c.Decode(&data)
	if err != nil {
		return nil, control.Err(err, "internal error", 1)
	}

	event, err := e.getEvent(data.ID)
	if err != nil {
		return nil, control.Err(err, "event does not exists", 2)
	}

	event.setBindState(data.State)
	return nil, nil
}

func (e *Events) callTriggerEvent(id string, data interface{}) error {
	data := api.TriggerEvent{
		ID: id,
	}

	return e.ctrl.OneShot("triggerEvent", &data)
}
