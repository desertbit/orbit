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
	"github.com/desertbit/orbit/pkg/transport"
)

type Hooks []Hook

// Hook allows third-party code to hook into orbit's logic, to implement for example
// logging or authentication functionality.
type Hook interface {
	// Close is called if the client closes.
	Close() error

	// OnSession is called if a new client session is connected to the service.
	// RPC and stream routines are handled after this hook.
	// Do not use the stream after returning from this hook.
	// Return an error to close the session and abort the initialization process.
	OnSession(s Session, stream transport.Stream) error

	// OnSessionClosed is called as soon as the session closes.
	OnSessionClosed(s Session)

	// OnCall is called before a call request.
	// Return an error to abort the call.
	OnCall(ctx Context, id string, callKey uint32) error

	// OnCallDone is called after a call request.
	// The context is the same as from the OnCall hook.
	// If err == nil, then the call completed successfully.
	OnCallDone(ctx Context, id string, callKey uint32, err error)

	// OnCallCanceled is called, if a call is canceled.
	// The context is the same as from the OnCall hook.
	OnCallCanceled(ctx Context, id string, callKey uint32)

	// OnStream is called during a new stream setup.
	// Return an error to abort the stream setup.
	OnStream(ctx Context, id string, stream transport.Stream) error
}
