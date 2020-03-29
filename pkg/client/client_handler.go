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

package client

import (
	"fmt"
	"runtime/debug"

	"github.com/desertbit/orbit/pkg/transport"
)

type clientHandler interface {
	hookClose() error
	hookOnSession(session Session, stream transport.Stream) error
	hookOnSessionClosed(session Session)

	hookOnCall(ctx Context, id string, callKey uint32) error
	hookOnCallDone(ctx Context, id string, callKey uint32, err error)
	hookOnCallCanceled(ctx Context, id string, callKey uint32)
	hookOnStream(ctx Context, id string) error
}

func (c *client) hookClose() (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookClose: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookClose: %v", e)
			}
		}
	}()

	// Call the Close hooks.
	for _, h := range c.hooks {
		err = h.Close()
		if err != nil {
			err = fmt.Errorf("close hook: %w", err)
			return
		}
	}
	return
}

func (c *client) hookOnSession(session Session, stream transport.Stream) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnSession: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnSession: %v", e)
			}
		}
	}()

	// Call the OnSession hooks.
	for _, h := range c.hooks {
		err = h.OnSession(session, stream)
		if err != nil {
			err = fmt.Errorf("on new session hook: %w", err)
			return
		}
	}
	return
}

func (c *client) hookOnSessionClosed(session Session) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnSessionClosed: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnSessionClosed: %v", e)
			}
		}
	}()

	// Call the OnSessionClosed hooks.
	for _, h := range c.hooks {
		h.OnSessionClosed(session)
	}
}

func (c *client) hookOnCall(ctx Context, id string, callKey uint32) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnCall: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnCall: %v", e)
			}
		}
	}()

	// Call the OnCall hooks.
	for _, h := range c.hooks {
		err = h.OnCall(ctx, id, callKey)
		if err != nil {
			err = fmt.Errorf("on call hook: %w", err)
			return
		}
	}
	return
}

func (c *client) hookOnCallDone(ctx Context, id string, callKey uint32, err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnCallDone: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnCallDone: %v", e)
			}
		}
	}()

	// Call the OnCallDone hooks.
	for _, h := range c.hooks {
		h.OnCallDone(ctx, id, callKey, err)
	}
}

func (c *client) hookOnCallCanceled(ctx Context, id string, callKey uint32) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnCallCanceled: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnCallCanceled: %v", e)
			}
		}
	}()

	// Call the OnCallCanceled hooks.
	for _, h := range c.hooks {
		h.OnCallCanceled(ctx, id, callKey)
	}
}

func (c *client) hookOnStream(ctx Context, id string) (err error) {
	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			err = ErrCatchedPanic
			if c.opts.PrintPanicStackTraces {
				c.log.Error().Msgf("catched panic: hookOnStream: %v\n%s", e, string(debug.Stack()))
			} else {
				c.log.Error().Msgf("catched panic: hookOnStream: %v", e)
			}
		}
	}()

	// Call the OnStream hooks.
	for _, h := range c.hooks {
		err = h.OnStream(ctx, id)
		if err != nil {
			err = fmt.Errorf("on stream hook: %w", err)
			return
		}
	}
	return
}
