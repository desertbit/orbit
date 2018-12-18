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

package orbit

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/signaler"
)

const (
	// Timeout for the opening of streams in the Init funcs.
	initOpenStreamTimeout = 20 * time.Second

	// Default key for the single control created inside the Init() func.
	defaultInitControl = "control"
	// Default key for the single signaler created inside the Init() func.
	defaultInitSignaler = "signaler"
)

// The InitAcceptStreams type is a map of AcceptStreamFunc, where the key
// is the id of the stream and the value the func that is used to accept it.
type InitAcceptStreams map[string]AcceptStreamFunc

// The InitControl type is used to initialize one control. It contains
// the functions that the remote peer can call and a config.
type InitControl struct {
	Funcs  control.Funcs
	Config *control.Config
}

// The InitControls type is a map, where the key is the id of a control
// and the value the associated InitControl.
type InitControls map[string]InitControl

// The InitSignal type is used to initialize one signal. It contains
// the id of the signal and a filter for it.
type InitSignal struct {
	ID     string
	Filter signaler.FilterFunc
}

// The InitSignaler type is used to initialize one signaler. It contains
// the signals that can be triggered and a config.
type InitSignaler struct {
	Signals []InitSignal
	Config  *control.Config
}

// The InitSignalers type is a map, where the key is the id of a signaler
// and the value the associated InitSignaler.
type InitSignalers map[string]InitSignaler

// The Init type is used during the initialization of the orbit session
// and contains the definition to accept streams and define exactly
// one control and one signaler.
type Init struct {
	AcceptStreams InitAcceptStreams
	Control       InitControl
	Signaler      InitSignaler
}

// The Init type is used during the initialization of the orbit session
// and contains the definition to accept streams and define many
// controls and many signalers.
type InitMany struct {
	AcceptStreams InitAcceptStreams
	Controls      InitControls
	Signalers     InitSignalers
}

// Init initializes the session by using InitMany(), but only defining one
// control and one signaler.
// If no more than one control/signaler are needed, this is the more convenient
// method to call.
// Ready() must be called manually for the control and signaler afterwards.
func (s *Session) Init(opts *Init) (
	control *control.Control,
	signaler *signaler.Signaler,
	err error,
) {
	controls, signalers, err := s.InitMany(&InitMany{
		AcceptStreams: opts.AcceptStreams,
		Controls: map[string]InitControl{
			defaultInitControl: opts.Control,
		},
		Signalers: map[string]InitSignaler{
			defaultInitSignaler: opts.Signaler,
		},
	})
	if err != nil {
		return
	}

	control = controls[defaultInitControl]
	signaler = signalers[defaultInitSignaler]
	return
}

// InitMany initializes this session. Pass nil to just start accepting streams.
// Ready() must be called manually for all controls and signaler afterwards.
func (s *Session) InitMany(opts *InitMany) (
	controls map[string]*control.Control,
	signalers map[string]*signaler.Signaler,
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

	// Register and initialize the signaler.
	var sigMutex sync.Mutex
	signalers = make(map[string]*signaler.Signaler)

	handleSignals := func(channel string, s *signaler.Signaler) {
		sigMutex.Lock()
		signalers[channel] = s
		sigMutex.Unlock()
	}

	for channel, sig := range opts.Signalers {
		s.openSignals(
			channel,
			sig.Signals, sig.Config,
			handleSignals, handleErr,
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
				handleErr(ErrOpenTimeout)
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

func (s *Session) openSignals(
	channel string,
	signals []InitSignal,
	config *control.Config,
	handleResult func(channel string, e *signaler.Signaler),
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
				handleErr(ErrOpenTimeout)
				return

			case <-s.CloseChan():
				handleErr(ErrClosed)
				return

			case stream = <-streamChan:
			}
		}

		// Create the signaler.
		sgnl := signaler.New(stream, config)
		for _, sig := range signals {
			if sig.Filter == nil {
				sgnl.AddSignal(sig.ID)
			} else {
				sgnl.AddSignalFilter(sig.ID, sig.Filter)
			}
		}

		// Close the signaler if the session closes.
		go func() {
			select {
			case <-s.CloseChan():
			case <-sgnl.CloseChan():
			}
			sgnl.Close()
		}()

		// Finally send the ready signaler to the handler.
		handleResult(channel, sgnl)
	}()
}
