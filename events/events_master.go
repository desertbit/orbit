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

func (e *Events) AddEvent(id string) {
	e.addEvent(id)
}

func (e *Events) AddEventFilter(id string, filterFunc FilterFunc) {
	event := e.addEvent(id)
	event.setFilterFunc(filterFunc)
}

func (e *Events) AddEvents(ids []string) {
	e.eventMapMutex.Lock()
	defer e.eventMapMutex.Unlock()

	for _, id := range ids {
		// Log if the event is overwitten.
		if _, ok := e.eventMap[id]; ok {
			e.logger.Printf("event '%s' registered more than once", id)
		}
		e.eventMap[id] = newEvent(id)
	}
}

func (e *Events) TriggerEvent(id string, data interface{}) (err error) {
	event, err := e.getEvent(id)
	if err != nil {
		return
	}

	if event.isActive() {
		// Check if the event is filtered out
		var conformsToFilter bool
		conformsToFilter, err = event.conformsToFilter(data)
		if err != nil || !conformsToFilter {
			return
		}

		err = e.callTriggerEvent(id, data)
		if err != nil {
			return
		}
	}

	return
}

//###############//
//### Private ###//
//###############//

func (e *Events) getEvent(id string) (event *event, err error) {
	e.eventMapMutex.Lock()
	event = e.eventMap[id]
	e.eventMapMutex.Unlock()

	if e == nil {
		err = ErrEventNotFound
	}
	return
}

func (e *Events) addEvent(id string) (ev *event) {
	ev = newEvent(id)

	e.eventMapMutex.Lock()
	defer e.eventMapMutex.Unlock()

	// Log if the event is overwitten.
	if _, ok := e.eventMap[id]; ok {
		e.logger.Printf("event '%s' registered more than once", id)
	}

	e.eventMap[id] = ev
	return
}

// Call the listeners on the remote peer.
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

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// Called if the remote peer wants to be informed about the given event.
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

	event.setActive(data.Active)
	return nil, nil
}

// Called when the remote peer wants to set a filter on an event.
func (e *Events) setEventFilter(ctx *control.Context) (v interface{}, err error) {
	var data api.SetEventFilter
	err = ctx.Decode(&data)
	if err != nil {
		err = control.Err(err, "internal error", 1)
		return
	}

	event, err := e.getEvent(data.ID)
	if err != nil {
		err = control.Err(err, "event does not exists", 2)
		return
	}

	err = event.setFilter(newContext(data.Data, e.codec))
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
