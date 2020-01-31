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

// Pass a nil yamux config for the default config.
func NewConn(conn net.Conn, conf *yamux.Config) (orbit.Conn, error) {
	return NewConnWithCloser(conn, conf, closer.New())
}

// Pass a nil yamux config for the default config.
func NewConnWithCloser(conn net.Conn, conf *yamux.Config, cl closer.Closer) (orbit.Conn, error) {
	return newSession(cl, conn, false, conf)
}

// Pass a nil yamux config for the default config.
func NewTCPConn(remoteAddr string, conf *yamux.Config) (orbit.Conn, error) {
	return NewTCPConnWithCloser(remoteAddr, conf, closer.New())
}

// Pass a nil yamux config for the default config.
func NewTCPConnWithCloser(remoteAddr string, conf *yamux.Config, cl closer.Closer) (orbit.Conn, error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}

	return NewConnWithCloser(conn, conf, cl)
}

// Pass a nil yamux config for the default config.
func NewTLSConn(remoteAddr string, tlsConf *tls.Config, conf *yamux.Config) (orbit.Conn, error) {
	return NewTLSConnWithCloser(remoteAddr, tlsConf, conf, closer.New())
}

// Pass a nil yamux config for the default config.
func NewTLSConnWithCloser(
	remoteAddr string,
	tlsConf *tls.Config,
	conf *yamux.Config,
	cl closer.Closer,
) (orbit.Conn, error) {
	conn, err := tls.Dial("tcp", remoteAddr, tlsConf)
	if err != nil {
		return nil, err
	}

	return NewConnWithCloser(conn, conf, cl)
}
