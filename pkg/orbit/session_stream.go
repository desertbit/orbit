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
	"github.com/desertbit/orbit/internal/packet"
)

func (s *Session) RegisterStream(id string, f StreamFunc) {
	s.streamFuncsMx.Lock()
	s.streamFuncs[id] = f
	s.streamFuncsMx.Unlock()
}

// OpenStream opens a new stream with the given channel ID.
func (s *Session) OpenStream(ctx context.Context, id string) (stream net.Conn, err error) {
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
		go func() {
			err := s.handleNewStream(stream)
			if err != nil {
				s.log.Error().
					Err(err).
					Msg("session: failed to handle new incoming stream")
			}
		}()
	}
}

func (s *Session) handleNewStream(stream net.Conn) (err error) {
	defer func() {
		// Catch panics. Might be caused by the channel interface.
		if e := recover(); e != nil {
			if s.cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: %v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: %v", e)
			}
		}

		// Close the stream on error.
		if err != nil {
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
		return fmt.Errorf("init stream header: %v", err)
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
		s.streamFuncsMx.Lock()
		f = s.streamFuncs[data.ID]
		s.streamFuncsMx.Unlock()
		if f == nil {
			return fmt.Errorf("stream handler for id '%s' does not exist", data.ID)
		}

		s.log.Debug().Str("id", data.ID).Msg("new raw stream")

		// Pass it the new stream.
		err = f(s, stream)
		if err != nil {
			return fmt.Errorf("stream='%v': %v", data.ID, err)
		}

	case api.StreamTypeCallAsync:
		// Pass the stream to the control.
		s.HandleCallAsync(stream)

	default:
		return fmt.Errorf("invalid stream type: %v", data.Type)
	}

	return
}
