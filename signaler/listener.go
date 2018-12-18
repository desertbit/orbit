/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package signaler

import "github.com/desertbit/closer"

const (
	// defaultLsChanSize is the default size of the listener's C channel.
	// This size determines how many events can be buffered, before writing
	// new events to the channel will block.
	// This value should be sufficiently high, to avoid blocking.
	defaultLsChanSize = 16
)

// The Listener type represents one party that is interested in being notified,
// when a specific signal gets triggered. It offers a public channel that
// contains the context of a trigger, each time it gets triggered.
type Listener struct {
	// C is filled up with the trigger actions from the signaler.
	// Use OffChan() to stop your reading routine.
	C <-chan *Context

	// c points to the same channel as C, but allows writing to it.
	// It is used internally to write the triggers onto the channel.
	c chan *Context

	// The id of this listener.
	id uint64
	// A flag whether we are interested in only exactly one trigger.
	once bool
	// A reference to the listeners type this listener is being handled by.
	ls *listeners

	// The closer that gets triggered if the listener should shutdown.
	closer closer.Closer
}

// newListener returns a new listener that is attached to the given
// listeners and has an event channel with the given size.
// The once flag indicates, whether this listener only wants to receive
// exactly one signal.
// Panics, when the channel's size is less than or equal to 0.
func newListener(ls *listeners, chanSize int, once bool) *Listener {
	// Panic, if an invalid channel size has been given.
	if chanSize <= 0 {
		panic("orbit: signal: invalid channel size for listener")
	}

	// Create the channel.
	c := make(chan *Context, chanSize)
	// Create the listener.
	l := &Listener{
		C:    c,
		c:    c,
		once: once,
		ls:   ls,
	}
	l.closer = closer.New(l.onClose)

	// Finally self-register to the given listeners.
	ls.add(l)

	return l
}

// OffChan is a public getter to the listener's closer channel.
// It allows users of this package to listen to the event, that their
// listener gets closed.
func (l *Listener) OffChan() <-chan struct{} {
	return l.closer.CloseChan()
}

// Off closes this listener, meaning that it will no longer receive triggers
// for the signal and gets removed from its listeners.
func (l *Listener) Off() {
	l.closer.Close()
}

//###############//
//### Private ###//
//###############//

// onClose is a func that should be executed when the listener shuts down.
// It signals to its listeners that it is no longer interested in receiving
// the signal and wants to be removed.
func (l *Listener) onClose() error {
	// Remove the listener from the listeners.
	l.ls.removeChan <- l.id
	return nil
}

// trigger triggers the listener with the provided request context.
// In case the listener has set the once flag, it gets removed immediately
// afterwards.
func (l *Listener) trigger(ctx *Context) {
	// If the listener is already closed, return.
	if l.closer.IsClosed() {
		return
	}

	// Write the trigger data to the channel.
	l.c <- ctx

	// If the listener is only interested in one trigger, remove it.
	if l.once {
		l.Off()
	}
}

// bindFunc binds the given trigger handler func to this listener
// by starting a routine that listens on the trigger channel (c) and
// executes the func whenever the listener gets triggered.
// When the handler func is no longer needed, Off() should be called
// on the listener to avoid a leaking goroutine.
func (l *Listener) bindFunc(f func(ctx *Context)) {
	go l.callFuncRoutine(f)
}

// callFuncRoutine takes a trigger handler func as argument to start
// a routine that listens for incoming triggers on the signal and executes
// the handler func for each of them.
// The routine gets closed, if the listener has the once flag set, or if
// the listener or its parent listeners are closed.
// Panics are recovered to avoid crashing the whole application when a panicking
// handler func is executed.
func (l *Listener) callFuncRoutine(f func(ctx *Context)) {
	// Recover panics.
	defer func() {
		if e := recover(); e != nil {
			l.ls.s.logger.Printf("listener func routine panic: %v", e)
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
			// The listener has been triggered, so execute the handler func
			// with the trigger context.
			f(ctx)

			// Close the listener, if it only wants to be notified once.
			if l.once {
				l.Off()
				return
			}
		}
	}
}
