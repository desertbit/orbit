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
	"errors"
	"fmt"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	initStreamTimeout        = 7 * time.Second
	asyncStreamMaxHeaderSize = 1024 // 1 KB
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

	var (
		n, bytesRead  int
		streamType    api.StreamType
		streamTypeBuf = make([]byte, 1)
	)

	// Set a read deadline for the header.
	err = stream.SetReadDeadline(time.Now().Add(initStreamTimeout))
	if err != nil {
		return
	}

	// Read the type from the stream.
	for bytesRead < 1 {
		n, err = stream.Read(streamTypeBuf[bytesRead:])
		if err != nil {
			return fmt.Errorf("init stream: failed to read stream type: %w", err)
		}
		bytesRead += n
	}
	streamType = api.StreamType(streamTypeBuf[0])

	// Decide the type of stream.
	switch streamType {
	case api.StreamTypeRaw:
		// Read the header from the stream.
		var header api.StreamRaw
		err = packet.ReadDecode(stream, &header, api.Codec, s.maxHeaderSize)
		if err != nil {
			return fmt.Errorf("init stream header: %w", err)
		}

		// Reset the deadlines.
		err = stream.SetDeadline(time.Time{})
		if err != nil {
			return
		}

		return s.handleRawStream(header.ID, header.Data, stream)

	case api.StreamTypeAsyncCall:
		// Read the header from the stream.
		var header api.StreamAsync
		err = packet.ReadDecode(stream, &header, api.Codec, asyncStreamMaxHeaderSize)
		if err != nil {
			return fmt.Errorf("init stream header: %w", err)
		}

		// Reset the deadlines.
		err = stream.SetDeadline(time.Time{})
		if err != nil {
			return
		}

		return s.handleAsyncCallStream(stream, header.ID)

	case api.StreamTypeCancelCalls:
		// Reset the deadlines.
		err = stream.SetDeadline(time.Time{})
		if err != nil {
			return
		}

		return s.handleCancelStream(stream)

	default:
		return fmt.Errorf("invalid stream type: %v", streamType)
	}
}

func (s *session) handleRawStream(id string, data map[string][]byte, stream transport.Stream) (err error) {
	// Get the stream.
	str, err := s.handler.getStream(id)
	if err != nil {
		return
	}

	// Create the service context.
	sctx := newContext(context.Background(), s, data)

	// Call the OnStream hooks.
	err = s.handler.hookOnStream(sctx, id)
	if err != nil {
		return fmt.Errorf("stream %s: %w", id, err)
	}

	if str.typ == streamTypeRaw {
		go func() {
			// Wait, until the stream is closed.
			select {
			case <-stream.ClosedChan():
			case <-s.ClosingChan():
			}

			// Call OnStreamClosed hooks.
			s.handler.hookOnStreamClosed(sctx, id, nil)
		}()

		s.handler.handleRawStream(sctx, str.f.(RawStreamFunc), stream)
		return
	}

	// Create the typed stream.
	ts := newTypedRWStream(stream, s.codec, s.maxArgSize, s.maxRetSize)

	// Call the hooks and function in a nested function.
	// We must pass the error from the function call to the done hook.
	err = func() (err error) {
		// Call OnStreamClosed hooks.
		defer func() {
			s.handler.hookOnStreamClosed(sctx, id, err)
		}()

		// Call the handler.
		return s.handler.handleTypedStream(sctx, ts, str.typ, str.f)
	}()

	if err != nil {
		// Check, if an orbit error was returned.
		var oErr Error
		if errors.As(err, &oErr) {
			retHeader.ErrCode = oErr.Code()
			retHeader.Err = oErr.Msg()
		}

		// Ensure an error message is always set.
		if retHeader.Err == "" {
			if s.sendInternalErrors {
				retHeader.Err = err.Error()
			} else {
				retHeader.Err = fmt.Sprintf("%s call failed", h.ID)
			}
		}

		ts.closeWithErr()



		// Reset the error, because we handled it already and the result should be send to the caller.
		err = nil
	}
}
