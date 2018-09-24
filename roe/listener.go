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

import "github.com/desertbit/closer"

const (
	defaultLsChanSize = 16
)

type Listener struct {
	// C is filled up with the triggered events.
	// Use OffChan() to stop your reading routine.
	C <-chan *Context
	c chan *Context

	id   uint64
	once bool
	ls   *listeners

	closer closer.Closer
}

func newListener(ls *listeners, chanSize int, once bool) *Listener {
	if chanSize <= 0 {
		panic("orbit: event: invalid channel size for listener")
	}

	c := make(chan *Context, chanSize)
	l := &Listener{
		C:    c,
		c:    c,
		once: once,
		ls:   ls,
	}
	l.closer = closer.New(l.onClose)

	// Finally self-register.
	ls.add(l)

	return l
}

func (l *Listener) onClose() error {
	// Remove the listener from the listeners.
	l.ls.removeChan <- l.id
	return nil
}

func (l *Listener) OffChan() <-chan struct{} {
	return l.closer.CloseChan()
}

func (l *Listener) Off() {
	l.closer.Close()
}

func (l *Listener) trigger(ctx *Context) {
	if l.closer.IsClosed() {
		return
	}

	l.c <- ctx

	if l.once {
		l.Off()
	}
}

func (l *Listener) bindFunc(f func(ctx *Context)) {
	go l.callFuncRoutine(f)
}

func (l *Listener) callFuncRoutine(f func(ctx *Context)) {
	defer func() {
		if e := recover(); e != nil {
			l.ls.e.logger.Printf("listener func routine panic: %v", e)
		}
	}()

	closeChan := l.closer.CloseChan()

	for {
		select {
		case <-l.ls.closeChan:
			return

		case <-closeChan:
			return

		case ctx := <-l.C:
			f(ctx)

			if l.once {
				l.Off()
				return
			}
		}
	}
}
