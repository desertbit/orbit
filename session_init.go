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

package orbit

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/events"
)

const (
	initOpenStreamTimeout = 12 * time.Second
)

type InitAcceptStreams map[string]AcceptStreamFunc

type InitControls map[string]struct {
	Funcs  control.Funcs
	Config *control.Config
}

type InitEvent struct {
	ID     string
	Filter events.FilterFunc
}

type InitEvents map[string]struct {
	Events []InitEvent
	Config *control.Config
}

type Init struct {
	AcceptStreams InitAcceptStreams
	Controls      InitControls
	Events        InitEvents
}

// Init initialized this session. Pass nil to just start accepting streams.
// Ready() must be called manually for all controls and events.
func (s *Session) Init(opts *Init) (
	controls map[string]*control.Control,
	ev map[string]*events.Events,
	err error,
) {
	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Just start the routines if no options are passed.
	if opts == nil {
		s.startRoutines()
		return
	}

	var (
		wg        sync.WaitGroup
		errorChan = make(chan error, 1)
	)

	handleErr := func(err error) {
		select {
		case errorChan <- err:
		default:
		}
	}

	// Register the accept handlers.
	for channel, f := range opts.AcceptStreams {
		s.AcceptStream(channel, f)
	}

	// Register and initialize the controls.
	var controlsMutex sync.Mutex
	controls = make(map[string]*control.Control)

	handleControl := func(channel string, ctrl *control.Control) {
		controlsMutex.Lock()
		controls[channel] = ctrl
		controlsMutex.Unlock()
	}

	for channel, c := range opts.Controls {
		s.openControl(
			channel, c.Funcs, c.Config,
			handleControl, handleErr,
			&wg, initOpenStreamTimeout,
		)
	}

	// Register and initialize the events.
	var evMutex sync.Mutex
	ev = make(map[string]*events.Events)

	handleEvents := func(channel string, e *events.Events) {
		evMutex.Lock()
		ev[channel] = e
		evMutex.Unlock()
	}

	for channel, e := range opts.Events {
		s.openEvents(
			channel,
			e.Events, e.Config,
			handleEvents, handleErr,
			&wg, initOpenStreamTimeout,
		)
	}

	// Start the session routines.
	// This will start accepting new streams.
	s.startRoutines()

	// Wait for all routines to finish.
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	// Wait for all goroutines or if an error occurs.
	select {
	case <-s.CloseChan():
		err = ErrClosed
		return

	case err = <-errorChan:
		return

	case <-waitChan:
	}

	// Ensure, that really no error happened.
	select {
	case err = <-errorChan:
		return
	default:
	}

	return
}

func (s *Session) openControl(
	channel string,
	funcs control.Funcs,
	config *control.Config,
	handleResult func(channel string, ctrl *control.Control),
	handleErr func(err error),
	wg *sync.WaitGroup,
	timeout time.Duration,
) {
	var (
		closeChan  = make(chan struct{})
		streamChan = make(chan net.Conn)
	)

	if !s.isClient {
		// This method must be called before startRoutines is called!
		s.AcceptStream(channel, func(conn net.Conn) error {
			select {
			case <-closeChan:
				conn.Close()
				return fmt.Errorf("not waiting for stream: accept disabled")
			case streamChan <- conn:
			}
			return nil
		})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(closeChan)

		var stream net.Conn

		if s.isClient {
			var err error
			stream, err = s.OpenStreamTimeout(channel, timeout)
			if err != nil {
				handleErr(err)
				return
			}
		} else {
			timeoutTimer := time.NewTimer(timeout)
			defer timeoutTimer.Stop()

			select {
			case <-timeoutTimer.C:
				handleErr(ErrTimeout)
				return

			case <-s.CloseChan():
				handleErr(ErrClosed)
				return

			case stream = <-streamChan:
			}
		}

		// Create the control.
		ctrl := control.New(stream, config)
		ctrl.AddFuncs(funcs)

		// Close the control if the session closes.
		go func() {
			select {
			case <-s.CloseChan():
			case <-ctrl.CloseChan():
			}
			ctrl.Close()
		}()

		// Finally send the ready control to the handler.
		handleResult(channel, ctrl)
	}()
}

func (s *Session) openEvents(
	channel string,
	evs []InitEvent,
	config *control.Config,
	handleResult func(channel string, e *events.Events),
	handleErr func(err error),
	wg *sync.WaitGroup,
	timeout time.Duration,
) {
	var (
		closeChan  = make(chan struct{})
		streamChan = make(chan net.Conn)
	)

	if !s.isClient {
		// This method must be called before startRoutines is called!
		s.AcceptStream(channel, func(conn net.Conn) error {
			select {
			case <-closeChan:
				conn.Close()
				return fmt.Errorf("not waiting for stream: accept disabled")
			case streamChan <- conn:
			}
			return nil
		})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(closeChan)

		var stream net.Conn

		if s.isClient {
			var err error
			stream, err = s.OpenStreamTimeout(channel, timeout)
			if err != nil {
				handleErr(err)
				return
			}
		} else {
			timeoutTimer := time.NewTimer(timeout)
			defer timeoutTimer.Stop()

			select {
			case <-timeoutTimer.C:
				handleErr(ErrTimeout)
				return

			case <-s.CloseChan():
				handleErr(ErrClosed)
				return

			case stream = <-streamChan:
			}
		}

		// Create the events.
		e := events.New(stream, config)
		for _, ev := range evs {
			if ev.Filter == nil {
				e.AddEvent(ev.ID)
			} else {
				e.AddEventFilter(ev.ID, ev.Filter)
			}
		}

		// Close the events if the session closes.
		go func() {
			select {
			case <-s.CloseChan():
			case <-e.CloseChan():
			}
			e.Close()
		}()

		// Finally send the ready events to the handler.
		handleResult(channel, e)
	}()
}
