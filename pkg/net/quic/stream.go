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
	"net"

	quic "github.com/lucas-clemente/quic-go"
)

var _ net.Conn = &stream{}

// Type stream wraps a quic.Stream and implements the
// two missing methods so it can be used as a net.Conn.
type stream struct {
	quic.Stream

	la net.Addr
	ra net.Addr
}

// newStream creates a new stream.
func newStream(qs quic.Stream, la, ra net.Addr) *stream {
	return &stream{
		Stream: qs,
		la:     la,
		ra:     ra,
	}
}

// Implements the net.Conn interface.
func (s *stream) LocalAddr() net.Addr {
	return s.la
}

// Implements the net.Conn interface.
func (s *stream) RemoteAddr() net.Addr {
	return s.ra
}
