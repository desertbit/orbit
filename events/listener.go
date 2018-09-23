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

import "sync"

const (
	defaultLsChanSize = 16
)

type Listener struct {
	C <-chan *Context

	ls *listeners

	id   uint64
	once bool
	c    chan *Context
	cMutex sync.Mutex
	cClosed bool

	closeChan <-chan struct{}
}

func newListener(ls *listeners, chanSize int, once bool, closeChan <-chan struct{}) *Listener {
	if chanSize <= 0 {
		panic("invalid channel size for listener")
	}

	c := make(chan *Context, chanSize)
	return &Listener{
		C:         c,
		ls:        ls,
		once:      once,
		c:         c,
		closeChan: closeChan,
	}
}

func (l *Listener) Off() {
	// Remove the listener from the listeners.
	l.ls.Remove(l.id)

	// Close the event channel. This ensures that any routines reading from
	// it get a chance to drain remaining events from it.
	l.cMutex.Lock()
	close(l.c)
	l.cClosed = true
	l.cMutex.Unlock()
}

func (l *Listener) handleEvent(ctx *Context) {
	l.cMutex.Lock()
	if !l.cClosed {
		l.c <- ctx
	}
	l.cMutex.Unlock()

	if l.once {
		l.Off()
	}
}

func (l *Listener) listenRoutine(f func(ctx *Context)) {
	for {
		select {
		case <-l.closeChan:
			break

		case ctx, more := <-l.C:
			if !more {
				return
			}

			f(ctx)
		}
	}
}
