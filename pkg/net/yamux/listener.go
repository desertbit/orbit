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

package yamux

import (
	"crypto/tls"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/hashicorp/yamux"
)

type listener struct {
	closer.Closer

	ln  net.Listener
	cfg *yamux.Config
}

func NewListener(ln net.Listener, cfg *yamux.Config) (orbit.Listener, error) {
	return NewListenerWithCloser(ln, cfg, closer.New())
}

func NewListenerWithCloser(ln net.Listener, cfg *yamux.Config, cl closer.Closer) (orbit.Listener, error) {
	l := &listener{
		Closer: cl,
		ln:     ln,
		cfg:    cfg,
	}
	l.OnClosing(ln.Close)
	return l, nil
}

func NewTCPListener(listenAddr string, cfg *yamux.Config) (orbit.Listener, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	return NewListener(ln, cfg)
}

func NewTLSListener(listenAddr string, tlsCfg *tls.Config, cfg *yamux.Config) (orbit.Listener, error) {
	ln, err := tls.Listen("tcp", listenAddr, tlsCfg)
	if err != nil {
		return nil, err
	}

	return NewListener(ln, cfg)
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (orbit.Conn, error) {
	c, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}

	return newSession(l.CloserOneWay(), c, true, l.cfg)
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.ln.Addr()
}
