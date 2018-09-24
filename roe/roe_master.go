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

func (r *ROE) AddEvent(id string) {
	r.addEvent(id)
}

func (r *ROE) AddEventFilter(id string, filterFunc FilterFunc) {
	event := r.addEvent(id)
	event.setFilterFunc(filterFunc)
}

func (r *ROE) AddEvents(ids []string) {
	r.eventsMutex.Lock()
	defer r.eventsMutex.Unlock()

	for _, id := range ids {
		// Log if the event is overwritten.
		if _, ok := r.events[id]; ok {
			r.logger.Printf("event '%s' registered more than once", id)
		}
		r.events[id] = newEvent(id)
	}
}

func (r *ROE) TriggerEvent(id string, data interface{}) (err error) {
	event, err := r.getEvent(id)
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

		err = r.callTriggerEvent(id, data)
		if err != nil {
			return
		}
	}

	return
}

//###############//
//### Private ###//
//###############//

func (r *ROE) getEvent(id string) (event *event, err error) {
	r.eventsMutex.Lock()
	event = r.events[id]
	r.eventsMutex.Unlock()

	if r == nil {
		err = ErrEventNotFound
	}
	return
}

func (r *ROE) addEvent(id string) (ev *event) {
	ev = newEvent(id)

	r.eventsMutex.Lock()
	defer r.eventsMutex.Unlock()

	// Log if the event is overwritten.
	if _, ok := r.events[id]; ok {
		r.logger.Printf("event '%s' registered more than once", id)
	}

	r.events[id] = ev
	return
}

// Call the listeners on the remote peer.
func (r *ROE) callTriggerEvent(id string, data interface{}) error {
	dataBytes, err := r.codec.Encode(data)
	if err != nil {
		return err
	}

	return r.roc.OneShot(cmdTriggerEvent, &api.TriggerEvent{
		ID:   id,
		Data: dataBytes,
	})
}

//###############################################//
//### Private - Callable from the remote Peer ###//
//###############################################//

// Called if the remote peer wants to be informed about the given event.
func (r *ROE) setEvent(c *roc.Context) (interface{}, error) {
	var data api.SetEvent
	err := c.Decode(&data)
	if err != nil {
		return nil, roc.Err(err, "internal error", 1)
	}

	event, err := r.getEvent(data.ID)
	if err != nil {
		return nil, roc.Err(err, "event does not exists", 2)
	}

	event.setActive(data.Active)
	return nil, nil
}

// Called when the remote peer wants to set a filter on an event.
func (r *ROE) setEventFilter(ctx *roc.Context) (v interface{}, err error) {
	var data api.SetEventFilter
	err = ctx.Decode(&data)
	if err != nil {
		err = roc.Err(err, "internal error", 1)
		return
	}

	event, err := r.getEvent(data.ID)
	if err != nil {
		err = roc.Err(err, "event does not exists", 2)
		return
	}

	err = event.setFilter(newContext(data.Data, r.codec))
	if err != nil {
		if err == ErrFilterFuncUndefined {
			err = roc.Err(err, "filter not set", 3)
		} else {
			err = roc.Err(err, "internal error", 1)
		}
		return
	}

	return
}
