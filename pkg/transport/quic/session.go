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
	"errors"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
	quic "github.com/quic-go/quic-go"
)

const (
	errorCodeClose = 0x1
)

var _ transport.Conn = &session{}

type session struct {
	closer.Closer

	qs quic.Connection
	la net.Addr
	ra net.Addr
}

func newSession(cl closer.Closer, qs quic.Connection) (s *session, err error) {
	s = &session{
		Closer: cl,
		qs:     qs,
		la:     qs.LocalAddr(),
		ra:     qs.RemoteAddr(),
	}
	s.OnClosing(func() error {
		return qs.CloseWithError(errorCodeClose, "closed")
	})

	// Always close on error.
	// Uncomment this, if code with errors is added!
	/*defer func() {
		if err != nil {
			s.Close_()
		}
	}()*/

	// Always close if the quic session closes.
	go func() {
		select {
		case <-s.ClosingChan():
		case <-qs.Context().Done():
		}
		s.Close_()
	}()

	return
}

// Implements the transport.Conn interface.
func (s *session) AcceptStream(ctx context.Context) (transport.Stream, error) {
	stream, err := s.qs.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	return newStream(stream, s.la, s.ra), nil
}

// Implements the transport.Conn interface.
func (s *session) OpenStream(ctx context.Context) (transport.Stream, error) {
	// We are using OpenStream instead of OpenStreamSync.
	// No context is required, because the stream is instantly opened.
	stream, err := s.qs.OpenStream()
	if err != nil {
		return nil, err
	}

	return newStream(stream, s.la, s.ra), nil
}

// Implements the transport.Conn interface.
func (s *session) LocalAddr() net.Addr {
	return s.la
}

// Implements the transport.Conn interface.
func (s *session) RemoteAddr() net.Addr {
	return s.ra
}

func (s *session) IsClosedError(err error) bool {
	var sErr *quic.StreamError
	if errors.As(err, &sErr) {
		return sErr.ErrorCode == errorCodeClose
	}
	return false
}
