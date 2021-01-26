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

package client

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

const (
	sessionHandshakeMaxPayloadSize = 1024 // 1 KB
)

type Session interface {
	closer.Closer

	ID() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type session struct {
	closer.Closer

	id      string
	conn    transport.Conn
	handler clientHandler
	log     *zerolog.Logger
	codec   codec.Codec
	chain   *chain

	maxArgSize    int
	maxRetSize    int
	maxHeaderSize int

	streamReadMx  sync.Mutex
	streamWriteMx sync.Mutex
	stream        transport.Stream

	cancelStreamMx sync.Mutex
	cancelStream   transport.Stream
}

// Implements the Session interface.
func (s *session) ID() string {
	return s.id
}

// Implements the Session interface.
func (s *session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// Implements the Session interface.
func (s *session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func connectSession(h clientHandler, opts Options) (s *session, err error) {
	ctxConnect, cancelConnect := context.WithTimeout(context.Background(), opts.ConnectTimeout)
	defer cancelConnect()

	// Connect to the service.
	conn, err := opts.Transport.Dial(opts.Closer.CloserOneWay(), ctxConnect, opts.TransportValue)
	if err != nil {
		return
	}

	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close_()
		}
	}()

	// Create new timeout context.
	ctx, cancel := context.WithTimeout(context.Background(), opts.HandshakeTimeout)
	defer cancel()

	// Open the stream to the server.
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		return
	}

	// Deadline is certainly available.
	deadline, _ := ctx.Deadline()

	// Set the deadline for the handshake.
	err = stream.SetDeadline(deadline)
	if err != nil {
		return
	}

	// Send the arguments for the handshake to the server.
	err = packet.WriteEncode(stream, &api.HandshakeArgs{Version: api.Version}, api.Codec, sessionHandshakeMaxPayloadSize)
	if err != nil {
		return
	}

	// Wait for the server's handshake response.
	var ret api.HandshakeRet
	err = packet.ReadDecode(stream, &ret, api.Codec, sessionHandshakeMaxPayloadSize)
	if err != nil {
		return
	} else if ret.Code == api.HSInvalidVersion {
		return nil, ErrInvalidVersion
	} else if ret.Code != api.HSOk {
		return nil, fmt.Errorf("unknown handshake code %d", ret.Code)
	}

	// Reset the deadline.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Finally, create the orbit session.
	s = &session{
		Closer: conn,

		id:      ret.SessionID,
		conn:    conn,
		handler: h,
		log:     opts.Log,
		codec:   opts.Codec,
		chain:   newChain(),

		maxArgSize:    opts.MaxArgSize,
		maxRetSize:    opts.MaxRetSize,
		maxHeaderSize: opts.MaxHeaderSize,

		stream: stream,
	}

	// Call the OnSession hooks.
	err = h.hookOnSession(s, stream)
	if err != nil {
		return
	}

	// Reset the deadlines. Might be set by some hooks.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Start the session routines.
	s.startRPCRoutines()

	// Call the OnSessionClosed hooks as soon as the session closes.
	s.OnClose(func() error {
		h.hookOnSessionClosed(s)
		return nil
	})

	return
}
