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

import "github.com/desertbit/orbit/control"

const (
	listenerDefaultChanSize = 16
)

type Listener struct {
	C <-chan *control.Context

	ls *listeners

	id uint64
	c  chan *control.Context
	once bool
}

func newListener(ls *listeners, chanSize int, once bool) *Listener {
	if chanSize <= 0 {
		panic("invalid channel size for listener")
	}

	c := make(chan *control.Context, chanSize)
	return &Listener{
		C:  c,
		ls: ls,
		c:  c,
		once: once,
	}
}

func (l *Listener) Off() {
	l.ls.Remove(l.id)
}
