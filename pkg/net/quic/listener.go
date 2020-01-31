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
	"crypto/tls"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/orbit"
	quic "github.com/lucas-clemente/quic-go"
)

var _ orbit.Listener = &listener{}

type listener struct {
	closer.Closer

	ln quic.Listener
}

func NewListener(ln quic.Listener) (orbit.Listener, error) {
	return NewListenerWithCloser(ln, closer.New())
}

// Pass a nil quic config for the default config.
func NewListenerWithCloser(ln quic.Listener, cl closer.Closer) (orbit.Listener, error) {
	l := &listener{
		Closer: cl,
		ln:     ln,
	}
	l.OnClosing(ln.Close)
	return l, nil
}

// Pass a nil quic config for the default config.
func NewTLSListener(listenAddr string, tlsConf *tls.Config, conf *quic.Config) (orbit.Listener, error) {
	return NewTLSListenerWithCloser(listenAddr, tlsConf, conf, closer.New())
}

// Pass a nil quic config for the default config.
func NewTLSListenerWithCloser(
	listenAddr string,
	tlsConf *tls.Config,
	conf *quic.Config,
	cl closer.Closer,
) (orbit.Listener, error) {
	// Prepare quic config.
	if conf == nil {
		conf = DefaultConfig()
	}

	// Create the listener.
	ln, err := quic.ListenAddr(listenAddr, tlsConf, conf)
	if err != nil {
		return nil, err
	}

	return NewListenerWithCloser(ln, cl)
}

// Implements the orbit.Listener interface.
func (l *listener) Accept() (orbit.Conn, error) {
	// No context passed, as listener is closed in onClosing of closer.
	qs, err := l.ln.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	return newSession(l.CloserOneWay(), qs)
}

// Implements the orbit.Listener interface.
func (l *listener) Addr() net.Addr {
	return l.ln.Addr()
}
