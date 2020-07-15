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
	"fmt"
	"runtime/debug"

	"github.com/desertbit/orbit/pkg/transport"
)

type serviceHandler interface {
	getCall(id string) (c call, err error)
	getAsyncCallOptions(id string) (opts asyncCallOptions, err error)
	getStream(id string) (str stream, err error)

	handleCall(ctx Context, f CallFunc, payload []byte) (ret interface{}, err error)
	handleRawStream(ctx Context, f RawStreamFunc, stream transport.Stream)
	handleTypedStream(ctx Context, str stream, stream transport.Stream) error

	hookClose() error
	hookOnSession(session Session, stream transport.Stream) error
	hookOnSessionClosed(session Session)

	hookOnCall(ctx Context, id string, callKey uint32) error
	hookOnCallDone(ctx Context, id string, callKey uint32, err error)
	hookOnCallCanceled(ctx Context, id string, callKey uint32)

	hookOnStream(ctx Context, id string) error
	hookOnStreamClosed(ctx Context, id string, err error)
}

func (s *service) getCall(id string) (c call, err error) {
	var ok bool
	c, ok = s.calls[id]
	if !ok {
		err = fmt.Errorf("call handler '%s' does not exist", id)
		return
	}
	return
}

func (s *service) getAsyncCallOptions(id string) (opts asyncCallOptions, err error) {
	var ok bool
	opts, ok = s.asyncCallOpts[id]
	if !ok {
		err = fmt.Errorf("no options available for async call '%s'", id)
	}
	return
}

func (s *service) getStream(id string) (str stream, err error) {
	var ok bool
	str, ok = s.streams[id]
	if !ok {
		err = fmt.Errorf("stream handler '%s' does not exist", id)
	}
	return
}

func (s *service) handleRawStream(ctx Context, f RawStreamFunc, stream transport.Stream) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: handleRawStream: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: handleRawStream: %v", e)
			}
		}
	}()

	f(ctx, stream)
}

func (s *service) handleTypedStream(ctx Context, str stream, stream transport.Stream) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: handleTypedStream: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: handleTypedStream: %v", e)
			}
		}
	}()

	ts := newTypedRWStream(stream, s.codec, str.maxArgSize, str.maxRetSize)
	switch str.typ {
	case streamTypeTR:
		err = str.f.(TypedRStreamFunc)(ctx, ts)
	case streamTypeTW:
		err = str.f.(TypedWStreamFunc)(ctx, ts)
	case streamTypeTRW:
		err = str.f.(TypedRWStreamFunc)(ctx, ts)
	default:
		return fmt.Errorf("stream type '%v' does not exist", str.typ)
	}
	return
}

func (s *service) handleCall(ctx Context, f CallFunc, payload []byte) (ret interface{}, err error) {
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

	// Execute the handler function.
	return f(ctx, payload)
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
			return fmt.Errorf("close hook: %w", err)
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
			return fmt.Errorf("on new session hook: %w", err)
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

func (s *service) hookOnCall(ctx Context, id string, callKey uint32) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnCall: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnCall: %v", e)
			}
		}
	}()

	// Call the OnCall hooks.
	for _, h := range s.hooks {
		err = h.OnCall(ctx, id, callKey)
		if err != nil {
			return fmt.Errorf("on call hook: %w", err)
		}
	}
	return
}

func (s *service) hookOnCallDone(ctx Context, id string, callKey uint32, err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnCallDone: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnCallDone: %v", e)
			}
		}
	}()

	// Call the OnCallDone hooks.
	for _, h := range s.hooks {
		h.OnCallDone(ctx, id, callKey, err)
	}
}

func (s *service) hookOnCallCanceled(ctx Context, id string, callKey uint32) {
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
		h.OnCallCanceled(ctx, id, callKey)
	}
}

func (s *service) hookOnStream(ctx Context, id string) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnStream: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnStream: %v", e)
			}
		}
	}()

	// Call the OnStream hooks.
	for _, h := range s.hooks {
		err = h.OnStream(ctx, id)
		if err != nil {
			return fmt.Errorf("on stream hook: %w", err)
		}
	}
	return
}

func (s *service) hookOnStreamClosed(ctx Context, id string, err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if s.opts.PrintPanicStackTraces {
				s.log.Error().Msgf("catched panic: hookOnStreamClosed: %v\n%s", e, string(debug.Stack()))
			} else {
				s.log.Error().Msgf("catched panic: hookOnStreamClosed: %v", e)
			}
		}
	}()

	// Call the OnStreamClosed hooks.
	for _, h := range s.hooks {
		h.OnStreamClosed(ctx, id, err)
	}
}
