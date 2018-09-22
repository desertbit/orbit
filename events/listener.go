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

const (
	listenerIDLen           = 16
	listenerDefaultChanSize = 16
)

type Listener struct {
	C <-chan *Context

	e  *Event
	id string
	c  chan *Context
}

func newListener(e *Event, chanSize int) (*Listener, error) {
	if chanSize <= 0 {
		return nil, ErrInvalidChanSize
	}

	c := make(chan *Context, chanSize)
	l := &Listener{
		C: c,
		e: e,
		c: c,
	}
	return l, nil
}

func (l *Listener) Off() {
	// TODO: Only remove once?!
	l.e.removeListener(l.id)
}
