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
	"io"
	"net"

	"github.com/desertbit/orbit/pkg/transport"
	quic "github.com/lucas-clemente/quic-go"
)

var _ transport.Stream = &stream{}

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

// Implements the transport.Stream interface.
func (s *stream) LocalAddr() net.Addr {
	return s.la
}

// Implements the transport.Stream interface.
func (s *stream) RemoteAddr() net.Addr {
	return s.ra
}

// Implements the transport.Stream interface.
func (s *stream) Read(b []byte) (n int, err error) {
	// We want the quic.Stream to behave like a net.Conn.
	// Since the quic.Stream implements the io.Reader interface, it may return
	// an io.EOF error also, if at least one byte could be read.
	// But we only want io.EOF, if the connection closed, which is the case, if no bytes
	// were read.
	n, err = s.Stream.Read(b)
	if n > 0 && err == io.EOF {
		// Ignore io.EOF, if at least one byte could be read.
		err = nil
	}
	return
}

// Implements the transport.Stream interface.
func (s *stream) Write(p []byte) (n int, err error) {
	// Check for the close error code from a CancelRead peer call.
	n, err = s.Stream.Write(p)
	if err != nil {
		if sErr, ok := err.(quic.StreamError); ok && sErr.ErrorCode() == errorCodeClose {
			err = io.EOF
		}
		return
	}
	return
}

// Implements the transport.Stream interface.
func (s *stream) Close() error {
	// Close the peer's writer, because Stream.Close does only a one way close.
	s.Stream.CancelRead(errorCodeClose)
	return s.Stream.Close()
}

// Implements the transport.Stream interface.
func (s *stream) IsClosed() bool {
	select {
	case <-s.Context().Done():
		return true
	default:
		return false
	}
}

// Implements the transport.Stream interface.
func (s *stream) ClosedChan() <-chan struct{} {
	return s.Context().Done()
}
