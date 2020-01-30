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

package orbit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
)

const (
	initStreamHeaderTimeout = 7 * time.Second
)

func (s *Session) RegisterStream(service, id string, f StreamFunc) {
	s.streamFuncsMx.Lock()
	s.streamFuncs[id] = f
	s.streamFuncsMx.Unlock()
}

// OpenStream opens a new stream with the given channel ID.
func (s *Session) OpenStream(ctx context.Context, service, id string) (stream net.Conn, err error) {
	return s.openStream(ctx, id, api.StreamTypeRaw)
}

//###############//
//### Private ###//
//###############//

func (s *Session) openStream(ctx context.Context, id string, t api.StreamType) (stream net.Conn, err error) {
	// Open the stream through our conn.
	stream, err = s.conn.OpenStream(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = stream.Close()
		}
	}()

	// Create the initial data that signals to the remote peer,
	// which stream we want to open.
	data := api.InitStream{
		ID:   id,
		Type: t,
	}

	// Set a write deadline, if needed.
	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		err = stream.SetWriteDeadline(deadline)
		if err != nil {
			return
		}
	}

	// Write the initial request to the stream.
	err = packet.WriteEncode(stream, &data, s.cf.Codec)
	if err != nil {
		return
	}

	// Reset the deadline.
	if hasDeadline {
		err = stream.SetWriteDeadline(time.Time{})
		if err != nil {
			return
		}
	}

	return
}

func (s *Session) acceptStreamRoutine() {
	defer s.Close_()

	ctx, cancel := context.WithCancel(context.Background())
	s.OnClosing(func() error {
		cancel()
		return nil
	})

	for {
		// Quit if the session has been closed.
		if s.IsClosing() {
			return
		}

		// Wait for new incoming connections.
		stream, err := s.conn.AcceptStream(ctx)
		if err != nil {
			if !s.IsClosing() && !errors.Is(err, io.EOF) {
				s.log.Error().
					Err(err).
					Msg("session: failed to accept stream")
			}
			return
		}

		// Run this in a new goroutine.
		go s.handleNewStream(stream)
	}
}

func (s *Session) handleNewStream(stream net.Conn) {
	var err error
	defer func() {
		// Catch panics. Might be caused by the channel interface.
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: \n%v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: \n%v", e)
			}
		}

		if err != nil {
			// Log. Do not use the Err() field, as stack trace formatting is lost then.
			s.log.Error().
				Msgf("session: failed to handle new incoming stream: \n%v", err)

			// Close the stream on error.
			_ = stream.Close()
		}
	}()

	// Set a read the deadline for the header.
	err = stream.SetReadDeadline(time.Now().Add(initStreamHeaderTimeout))
	if err != nil {
		return
	}

	// Read the initial data from the stream.
	var data api.InitStream
	err = packet.ReadDecode(stream, &data, s.cf.Codec)
	if err != nil {
		err = fmt.Errorf("init stream header: %v", err)
		return
	}

	// Reset the deadline.
	err = stream.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}

	// Decide the type of stream.
	switch data.Type {
	case api.StreamTypeRaw:
		// Obtain the stream handler.
		var f StreamFunc
		s.streamFuncsMx.RLock()
		f = s.streamFuncs[data.ID]
		s.streamFuncsMx.RUnlock()
		if f == nil {
			err = fmt.Errorf("stream handler for id '%s' does not exist", data.ID)
			return
		}

		// Authorize the stream, if needed.
		if s.authz != nil && !s.authz(s, data.ID) {
			err = fmt.Errorf("unauthorized access to stream '%s'", data.ID)
			return
		}

		s.log.Debug().Str("id", data.ID).Msg("new raw stream")

		// Pass it the new stream.
		// The stream must be closed by the handler!
		f(s, stream)

	case api.StreamTypeCallAsync:
		s.log.Debug().Msg("new call async stream")

		// Handle the new async call stream.
		s.handleAsyncCall(stream)

	case api.StreamTypeCallInit:
		s.log.Debug().Msg("new call init stream")

		// Handle the new init call stream.
		s.handleInitCall(stream)

	default:
		err = fmt.Errorf("invalid stream type: %v", data.Type)
		return
	}
}
