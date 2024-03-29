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

package yamux

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
)

type yTransport struct {
	opts *Options
}

func NewTransport(opts *Options) (t transport.Transport, err error) {
	// Set the default options.
	opts.setDefaults()

	// Validate the options.
	err = opts.validate()
	if err != nil {
		return
	}

	t = &yTransport{opts: opts}

	return
}

func (t *yTransport) Dial(cl closer.Closer, ctx context.Context, addr string) (tc transport.Conn, err error) {
	// Open the connection.
	var conn net.Conn
	if t.opts.TLSConfig != nil {
		// TODO: Cancel is deprecated, remove once https://github.com/golang/go/issues/18482 is merged and replace with DialContext.
		dl := &net.Dialer{Cancel: ctx.Done()}
		conn, err = tls.DialWithDialer(dl, t.opts.Network, addr, t.opts.TLSConfig)
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, t.opts.Network, addr)
	}
	if err != nil {
		return
	}

	return newSession(cl, conn, false, t.opts.Config)
}

func (t *yTransport) Listen(cl closer.Closer, addr string) (tl transport.Listener, err error) {
	// Create the listener.
	var ln net.Listener
	if t.opts.TLSConfig != nil {
		ln, err = tls.Listen(t.opts.Network, addr, t.opts.TLSConfig)
	} else {
		ln, err = net.Listen(t.opts.Network, addr)
	}
	if err != nil {
		return
	}

	return newListener(cl, ln, t.opts.Config), nil
}
