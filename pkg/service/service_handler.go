/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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

package service

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/desertbit/orbit/pkg/transport"
)

type serviceHandler interface {
	handleStream(session Session, id string, data map[string][]byte, stream transport.Stream) error
	handleCall(sctx Context, id string, callKey uint32, payload []byte) (ret interface{}, err error)

	hookOnCallCanceled(sctx Context, id string, callKey uint32)
	hookClose() error
	hookOnSession(session Session, stream transport.Stream) error
	hookOnSessionClosed(session Session)
}

func (s *service) handleStream(session Session, id string, data map[string][]byte, stream transport.Stream) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: handleStream: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: handleStream: %v", e)
			}
		}
	}()

	// Obtain the stream handler function.
	f, ok := s.streamFunc(id)
	if !ok {
		return fmt.Errorf("stream handler '%s' does not exist", id)
	}

	// Create the service context.
	sctx := newContext(context.Background(), session, data)

	// Call the OnStream hooks.
	for _, h := range s.hooks {
		err = h.OnStream(sctx, id, stream)
		if err != nil {
			err = fmt.Errorf("on new stream hook: %w", err)
			return
		}
	}

	// Pass the new stream.
	// The stream must be closed by the handler!
	f(sctx, stream)

	return
}

func (s *service) handleCall(sctx Context, id string, callKey uint32, payload []byte) (ret interface{}, err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: handleCall: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: handleCall: %v", e)
			}
		}
	}()

	// Obtain the call function.
	f, ok := s.callFunc(id)
	if !ok {
		err = fmt.Errorf("call handler '%s' does not exist", id)
		return
	}

	// Call the OnCall hooks.
	for _, h := range s.hooks {
		err = h.OnCall(sctx, id, callKey)
		if err != nil {
			err = fmt.Errorf("on call hook: %w", err)
			return
		}
	}

	// Execute the handler function.
	ret, err = f(sctx, payload)

	// Call the OnCallDone hooks.
	for _, h := range s.hooks {
		h.OnCallDone(sctx, id, callKey, err)
	}
	return ret, err
}

func (s *service) hookOnCallCanceled(sctx Context, id string, callKey uint32) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnCallCanceled: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnCallCanceled: %v", e)
			}
		}
	}()

	// Call the OnCallCanceled hooks.
	for _, h := range s.hooks {
		h.OnCallCanceled(sctx, id, callKey)
	}
}

func (s *service) hookClose() (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookClose: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookClose: %v", e)
			}
		}
	}()

	// Call the Close hooks.
	for _, h := range s.hooks {
		err = h.Close()
		if err != nil {
			err = fmt.Errorf("close hook: %w", err)
			return
		}
	}
	return
}

func (s *service) hookOnSession(session Session, stream transport.Stream) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnSession: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnSession: %v", e)
			}
		}
	}()

	// Call the OnSession hooks.
	for _, h := range s.hooks {
		err = h.OnSession(session, stream)
		if err != nil {
			err = fmt.Errorf("on new session hook: %w", err)
			return
		}
	}
	return
}

func (s *service) hookOnSessionClosed(session Session) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnSessionClosed: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnSessionClosed: %v", e)
			}
		}
	}()

	// Call the OnSessionClosed hooks.
	for _, h := range s.hooks {
		h.OnSessionClosed(session)
	}
}
