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
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
	quic "github.com/quic-go/quic-go"
)

var _ transport.Listener = &listener{}

type listener struct {
	closer.Closer

	ln *quic.Listener
}

func newListener(cl closer.Closer, ln *quic.Listener) transport.Listener {
	l := &listener{
		Closer: cl,
		ln:     ln,
	}
	l.OnClosing(ln.Close)
	return l
}

// Implements the transport.Listener interface.
func (l *listener) Accept() (transport.Conn, error) {
	// No context passed, as listener is closed in onClosing of closer.
	qs, err := l.ln.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	return newSession(l.CloserOneWay(), qs)
}

// Implements the transport.Listener interface.
func (l *listener) Addr() net.Addr {
	return l.ln.Addr()
}
