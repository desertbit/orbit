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

import (
	"sync"
)

// The Group type represents a group of signalers.
// It is useful for triggering the same signal with some data
// on all signalers contained in the group.
// A Group is thread-safe.
type Group struct {
	// Synchronises the access to the signalers.
	mutex sync.RWMutex
	// The signalers that are part of this group.
	signalers []*Signaler
}

// NewGroup creates an empty group.
func NewGroup() *Group {
	return &Group{
		signalers: make([]*Signaler, 0),
	}
}

// Add adds the given signalers to the group.
// When a signaler is closed, it is automatically
// removed from the group. There is no need to call Remove()
// in that case.
func (g *Group) Add(ss ...*Signaler) {
	if len(ss) == 0 {
		return
	}

	g.mutex.Lock()
	for _, s := range ss {
		// Ignore nil signalers.
		if s == nil {
			continue
		}

		// Add the signaler to our group.
		g.signalers = append(g.signalers, s)

		// Remove the signaler from the group once it closes.
		// ATTENTION! The reference to the signaler must be stored in a variable here!
		// Otherwise, the pointer of the for loop is used and that is overwritten
		// in each iteration, leading to all signalers closing the last signaler.
		refToS := s
		s.OnClose(func() error {
			g.Remove(refToS)
			return nil
		})
	}
	g.mutex.Unlock()
}

// Remove removes the given signalers from this group.
// If a signaler is not in the group, this method is a no-op for it.
func (g *Group) Remove(ss ...*Signaler) {
	g.mutex.Lock()

Loop:
	for _, s := range ss {
		for i := 0; i < len(g.signalers); i++ {
			if g.signalers[i] == s {
				// Delete without causing a memory leak due to the gc
				// not being able to collect the pointer. This is done
				// by explicitly overwriting the element.
				// We DO NOT preserve the original order of the slice,
				// as this is not important to us.
				// See: https://github.com/golang/go/wiki/SliceTricks
				l := len(g.signalers)
				g.signalers[i] = g.signalers[l-1]
				g.signalers[l-1] = nil
				g.signalers = g.signalers[:l-1]
				continue Loop
			}
		}
	}
	g.mutex.Unlock()
}

// TriggerSignal calls each signaler's TriggerSignal() method of this group
// with the given id and the given data.
// If a signaler in the group does not contain the signal with the given id,
// the error is ignored and all other signalers are still triggered.
//
// Signalers can be explicitly excluded from being triggered.
func (g *Group) TriggerSignal(id string, data interface{}, exclude ...*Signaler) error {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	// Trigger the signal for all signalers.
TriggerLoop:
	for _, s := range g.signalers {
		// Do not trigger the signal for the excluded signalers.
		for _, ex := range exclude {
			if ex == s {
				continue TriggerLoop
			}
		}

		// It is ignored, if one of the signalers does not contain the desired signal.
		err := s.TriggerSignal(id, data)
		if err != nil && err != ErrSignalNotFound {
			return err
		}
	}

	return nil
}