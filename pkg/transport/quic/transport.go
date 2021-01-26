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

package quic

import (
	"context"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/options"
	"github.com/desertbit/orbit/pkg/transport"
	quic "github.com/lucas-clemente/quic-go"
)

type qTransport struct {
	opts Options
}

func NewTransport(opts Options) (t transport.Transport, err error) {
	// Set default values, where needed.
	err = options.SetDefaults(&opts, DefaultOptions("", "", nil))
	if err != nil {
		return
	}

	// Validate the options.
	err = opts.validate()
	if err != nil {
		return
	}

	t = &qTransport{opts: opts}

	return
}

// Implements the transport.Transport interface.
func (q *qTransport) Dial(cl closer.Closer, ctx context.Context, v interface{}) (transport.Conn, error) {
	// Create a quic connection.
	qs, err := quic.DialAddrContext(ctx, q.opts.DialAddr, q.opts.TLSConfig, q.opts.Config)
	if err != nil {
		return nil, err
	}

	// Create a new session.
	return newSession(cl, qs)
}

// Implements the transport.Transport interface.
func (q *qTransport) Listen(cl closer.Closer, v interface{}) (transport.Listener, error) {
	// Create the quic listener.
	ln, err := quic.ListenAddr(q.opts.ListenAddr, q.opts.TLSConfig, q.opts.Config)
	if err != nil {
		return nil, err
	}

	// Create a new listener.
	return newListener(cl, ln), nil
}
