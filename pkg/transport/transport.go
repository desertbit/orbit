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

package transport

import (
	"context"
	"net"

	"github.com/desertbit/closer/v3"
)

type Transport interface {
	Dial(cl closer.Closer, ctx context.Context, addr string) (Conn, error)
	Listen(cl closer.Closer, addr string) (Listener, error)
}

type Conn interface {
	closer.Closer

	// AcceptStream returns the next stream opened by the peer, blocking until one is available.
	AcceptStream(context.Context) (Stream, error)

	// OpenStream opens a new bidirectional stream.
	// There is no signaling to the peer about new streams.
	// The peer can only accept the stream after data has been sent on it.
	OpenStream(context.Context) (Stream, error)

	// LocalAddr returns the local address.
	LocalAddr() net.Addr

	// RemoteAddr returns the address of the peer.
	RemoteAddr() net.Addr

	// IsClosedError checks whenever the passed error is a closed connection error.
	IsClosedError(error) bool
}

type Stream interface {
	net.Conn

	// IsClosed returns true, if the stream has been closed locally or by the remote peer.
	IsClosed() bool

	// ClosedChan returns a closed channel as soon as the stream closes.
	ClosedChan() <-chan struct{}
}

type Listener interface {
	closer.Closer

	// Accept waits for and returns the next connection to the listener.
	// The listener must close the new connection if the listener is closed.
	Accept() (Conn, error)

	// Addr returns the listener's network address.
	Addr() net.Addr
}
