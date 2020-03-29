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

package service

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

	id                 string
	conn               transport.Conn
	handler            serviceHandler
	codec              codec.Codec
	log                *zerolog.Logger
	sendInternalErrors bool

	maxArgSize    int
	maxRetSize    int
	maxHeaderSize int

	streamWriteMx sync.Mutex
	stream        transport.Stream

	cancelMx        sync.Mutex
	cancelCalls     map[uint32]context.CancelFunc
	hasCancelStream bool
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

// Initialize the session and perform a handshake.
func initSession(
	conn transport.Conn,
	id string,
	h serviceHandler,
	opts *Options,
) (s *session, err error) {
	// Close connection on error.
	defer func() {
		if err != nil {
			conn.Close_()
		}
	}()

	// Create new context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), opts.HandshakeTimeout)
	defer cancel()

	// Wait for an incoming stream from the client.
	stream, err := conn.AcceptStream(ctx)
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

	// Wait for the client's handshake request.
	var args api.HandshakeArgs
	err = packet.ReadDecode(stream, &args, api.Codec, sessionHandshakeMaxPayloadSize)
	if err != nil {
		return
	}

	// Set the response data for the handshake.
	var ret api.HandshakeRet
	if args.Version != api.Version {
		ret.Code = api.HSInvalidVersion
	} else {
		ret.Code = api.HSOk
		ret.SessionID = id
	}

	// Send the handshake response back to the client.
	err = packet.WriteEncode(stream, &ret, api.Codec, sessionHandshakeMaxPayloadSize)
	if err != nil {
		return
	}

	// Abort, if the handshake failed.
	if ret.Code == api.HSInvalidVersion {
		err = ErrInvalidVersion
		return
	} else if ret.Code != api.HSOk {
		err = fmt.Errorf("unknown handshake code %d", ret.Code)
		return
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Create the session.
	s = &session{
		Closer: conn,

		id:                 id,
		conn:               conn,
		handler:            h,
		codec:              opts.Codec,
		log:                opts.Log,
		sendInternalErrors: opts.SendInternalErrors,

		maxArgSize:    opts.MaxArgSize,
		maxRetSize:    opts.MaxRetSize,
		maxHeaderSize: opts.MaxHeaderSize,

		stream: stream,

		cancelCalls: make(map[uint32]context.CancelFunc),
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
	s.startRPCReadRoutine()
	s.startAcceptStreamRoutine()

	// Call the OnSessionClosed hooks as soon as the session closes.
	s.OnClose(func() error {
		h.hookOnSessionClosed(s)
		return nil
	})

	return
}
