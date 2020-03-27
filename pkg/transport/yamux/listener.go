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
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/hashicorp/yamux"
)

var _ transport.Listener = &listener{}

type listener struct {
	closer.Closer

	ln   net.Listener
	conf *yamux.Config
}

func newListener(cl closer.Closer, ln net.Listener, conf *yamux.Config) transport.Listener {
	l := &listener{
		Closer: cl,
		ln:     ln,
		conf:   conf,
	}
	l.OnClosing(ln.Close)
	return l
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (transport.Conn, error) {
	c, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}

	return newSession(l.CloserOneWay(), c, true, l.conf)
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.ln.Addr()
}
