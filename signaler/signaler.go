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

/*
Package signaler makes it possible to send events between connected peers
in a network. It uses the orbit control package to send and register
these events (https://github.com/desertbit/orbit/control).

The signals can carry a payload.

Registering signals

Each peer can register signals during the initialization, which are then
available to the remote peer to listen to. A signal must have a unique
identifier.

Listening on signals

The remote peer can then listen on the registered signals by using either
the On- functions (getting notified every time the signal is triggered) or
the Once- functions (getting notified only once).
At each point in time, a remote peer can also stop listening on signals
to not receive any notifications after that.

Filtering

It is possible when adding a new signal to set a filter function on a signal.
Peers can then set filter data on this signal that gets passed to this filter
func. Depending on the return of the filter func, the signal may not be
triggered for this peer, if its filter is not fulfilled.
This allows peers to filter the signals to their liking and prevents a
waste of resources.
*/
package signaler

import (
	"log"
	"net"
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/control"
)

const (
	// The id for the control func that sets a signal.
	cmdSetSignalState = "SetSignalState"
	// The id for the control func that triggers a signal.
	cmdTriggerSignal = "TriggerSignal"
	// The id for the control func that sets a filter on the signal.
	cmdSetSignalFilter = "SetSignalFilter"
)

// The Signaler type is the main type to interact with signals
// on one peer. It contains the underlying control.Control,
// codecs, loggers, etc.
// It takes care of storing signals that have been added to it
// and keeps track of who is listening on which signal.
type Signaler struct {
	closer.Closer

	// The underlying control used to send the events on.
	ctrl *control.Control
	// The codec used to encode the payloads of the signals.
	codec codec.Codec
	// The logger used to log messages to.
	logger *log.Logger

	// Synchronises the access to the signals.
	signalsMutex sync.Mutex
	// Stores the signals that have been added to the signaler.
	// The key is the id of the signal.
	signals map[string]*signal

	// Synchronises the access to the listeners.
	lsMapMutex sync.Mutex
	// Stores the listeners for each signal. The key is the
	// id of the respective signal the listeners are interested
	// in.
	lsMap map[string]*listeners
}

// New returns a new Signaler.
func New(conn net.Conn, config *control.Config) (s *Signaler) {
	// Create the control.
	ctrl := control.New(conn, config)

	// Create the signaler.
	s = &Signaler{
		Closer:  ctrl,
		ctrl:    ctrl,
		codec:   ctrl.Codec(),
		logger:  ctrl.Logger(),
		signals: make(map[string]*signal),
		lsMap:   make(map[string]*listeners),
	}

	// Add the functions needed for the signaling to the control.
	s.ctrl.AddFuncs(control.Funcs{
		cmdSetSignalState:  s.setSignalState,
		cmdTriggerSignal:   s.triggerSignal,
		cmdSetSignalFilter: s.setSignalFilter,
	})
	return
}

// Ready signalizes that the initialization is done.
// Signaler can now be triggered.
// This should be only called once.
func (s *Signaler) Ready() {
	s.ctrl.Ready()
}
