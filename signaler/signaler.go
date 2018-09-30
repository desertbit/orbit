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

It is possible for a remote peer to set a filter on a signal. Right now,
this filter is then applied for the signal in general, meaning that all
peers listening on the same signal are affected by the same filter.
A filter allows to configure when exactly a signal may be triggered.
*/
package signaler

import (
	"log"
	"net"
	"sync"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/control"

	"github.com/desertbit/closer"
)

const (
	// The id for the control func that sets a signal.
	cmdSetSignal       = "SetSignal"
	// The id for the control func that triggers a signal.
	cmdTriggerSignal   = "TriggerSignal"
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
	ctrl   *control.Control
	// The codec used to encode the payloads of the signals.
	codec  codec.Codec
	// The logger used to log messages to.
	logger *log.Logger

	// Synchronises the access to the signals.
	signalsMutex sync.Mutex
	// Stores the signals that have been added to the signaler.
	// The key is the id of the signal.
	signals      map[string]*signal

	// Synchronises the access to the listeners.
	lsMapMutex sync.Mutex
	// Stores the listeners for each signal. The key is the
	// id of the respective signal the listeners are interested
	// in.
	lsMap      map[string]*listeners
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
		cmdSetSignal:       s.setSignal,
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
