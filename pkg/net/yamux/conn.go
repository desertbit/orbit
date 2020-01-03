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

func NewConn(conn net.Conn, cfg *yamux.Config) (orbit.Conn, error) {
	return NewConnWithCloser(conn, cfg, closer.New())
}

func NewConnWithCloser(conn net.Conn, cfg *yamux.Config, cl closer.Closer) (orbit.Conn, error) {
	return newSession(cl, conn, false, cfg)
}

func NewTCPConn(remoteAddr string, cfg *yamux.Config) (orbit.Conn, error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}

	return NewConn(conn, cfg)
}

func NewTLSConn(remoteAddr string, tlsCfg *tls.Config, cfg *yamux.Config) (orbit.Conn, error) {
	conn, err := tls.Dial("tcp", remoteAddr, tlsCfg)
	if err != nil {
		return nil, err
	}

	return NewConn(conn, cfg)
}
