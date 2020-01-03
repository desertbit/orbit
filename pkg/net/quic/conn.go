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

package quic

import (
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/orbit"
)

func NewConn(
	conn net.PacketConn,
	remoteAddr net.Addr,
	host string,
	tlsConf *tls.Config,
	conf *quic.Config,
	cl closer.Closer,
) (orbit.Conn, error) {
	return NewConnWithCloser(conn, remoteAddr, tlsConf, conf, closer.New())
}

func NewConnWithCloser(
	conn net.PacketConn,
	remoteAddr net.Addr,
	host string,
	tlsConf *tls.Config,
	conf *quic.Config,
	cl closer.Closer,
) (orbit.Conn, error) {
	qs, err := quic.Dial(conn, remoteAddr, host, tlsConf, conf)
	if err != nil {
		return nil, err
	}

	return newSession(qs)
}
