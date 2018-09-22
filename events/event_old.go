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

/*
type Event struct {
	id string
	//filter func(ctx *Context)

	listenersMutex sync.Mutex
	listeners      map[string]*Listener
}

func newEvent(id string) *Event {
	return &Event{
		id: id,
	}
}

func (e *Event) On() (*Listener, error) {
	l, err := newListener(e, defaultListenerChanSize)
	if err != nil {
		return nil, err
	}

	// Create a new listener ID and ensure it is unqiue.
	// Add it to the listeners map and set the ID.
	for {
		l.id, err = utils.RandomString(listenerIDLen)
		if err != nil {
			return nil, err
		}

		added := func() bool {
			e.listenersMutex.Lock()
			defer e.listenersMutex.Unlock()

			if _, ok := c.listeners[l.id]; ok {
				return false
			}

			c.listeners[l.id] = l
			return true
		}()
		if added {
			break
		}
	}

	return l, nil
}

func (e *Event) removeListener(id string) {
	listenersMutex.Lock()
	delete(e.listeners, id)
	listenersMutex.Unlock()
}
*/
