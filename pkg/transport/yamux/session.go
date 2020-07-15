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
	"context"
	"errors"
	"io"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/desertbit/yamux"
)

var _ transport.Conn = &session{}

type session struct {
	closer.Closer

	conn net.Conn
	ys   *yamux.Session
}

func newSession(cl closer.Closer, conn net.Conn, isServer bool, conf *yamux.Config) (s *session, err error) {
	s = &session{
		Closer: cl,
		conn:   conn,
	}
	s.OnClosing(conn.Close)

	// Always close on error.
	defer func() {
		if err != nil {
			s.Close_()
		}
	}()

	// Create a new yamux session.
	if isServer {
		s.ys, err = yamux.Server(conn, conf)
	} else {
		s.ys, err = yamux.Client(conn, conf)
	}
	if err != nil {
		return
	}
	s.OnClosing(s.ys.Close)

	// Always close if the yamux session closes.
	go func() {
		select {
		case <-s.ClosingChan():
		case <-s.ys.ClosedChan():
		}
		s.Close_()
	}()

	return s, nil
}

// LocalAddr returns the local address.
func (s *session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the address of the peer.
func (s *session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// AcceptStream returns the next stream opened by the peer, blocking until one is available.
func (s *session) AcceptStream(ctx context.Context) (transport.Stream, error) {
	// TODO: Fork the yamux package and implement the context cancel handling.
	stream, err := s.ys.AcceptStream()
	if err != nil {
		return nil, err
	}

	return newStream(stream), nil
}

// OpenStream opens a new bidirectional stream.
// There is no signaling to the peer about new streams:
// The peer can only accept the stream after data has been sent on the stream.
func (s *session) OpenStream(ctx context.Context) (transport.Stream, error) {
	// TODO: Fork the yamux package and implement the context cancel handling.
	stream, err := s.ys.OpenStream()
	if err != nil {
		return nil, err
	}

	return newStream(stream), nil
}

func (s *session) IsClosedError(err error) bool {
	return errors.Is(err, io.EOF)
}
