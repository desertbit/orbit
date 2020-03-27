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
	"fmt"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	initStreamHeaderTimeout = 7 * time.Second
)

func (s *session) startAcceptStreamRoutine() {
	go s.acceptStreamRoutine()
}

func (s *session) acceptStreamRoutine() {
	defer s.Close_()

	// Create a new context from the closer.
	ctx, cancel := s.Closer.Context()
	defer cancel()

	for {
		// Quit if the session has been closed.
		if s.IsClosing() {
			return
		}

		// Wait for new incoming connections.
		stream, err := s.conn.AcceptStream(ctx)
		if err != nil {
			if !s.IsClosing() && !s.conn.IsClosing() && !s.conn.IsClosedError(err) {
				s.log.Error().
					Err(err).
					Msg("session: failed to accept stream")
			}
			return
		}

		// Run this in a new goroutine.
		go func() {
			gerr := s.handleNewStream(stream)
			if gerr != nil {
				s.log.Error().
					Err(gerr).
					Msg("session: failed to handle new incoming stream")
			}
		}()
	}
}

func (s *session) handleNewStream(stream transport.Stream) (err error) {
	// Close the stream on error.
	defer func() {
		if err != nil {
			_ = stream.Close()
		}
	}()

	// Set a read deadline for the header.
	err = stream.SetReadDeadline(time.Now().Add(initStreamHeaderTimeout))
	if err != nil {
		return
	}

	// Read the header from the stream.
	var header api.InitStream
	err = packet.ReadDecode(stream, &header, api.Codec)
	if err != nil {
		return fmt.Errorf("init stream header: %w", err)
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Decide the type of stream.
	switch header.Type {
	case api.StreamTypeRaw:
		return s.handler.handleStream(s, header.ID, header.Data, stream)

	case api.StreamTypeAsyncCall:
		return s.handleAsyncCallStream(stream)

	case api.StreamTypeCancelCalls:
		return s.handleCancelStream(stream)

	default:
		return fmt.Errorf("invalid stream type: %v", header.Type)
	}
}
