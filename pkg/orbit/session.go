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
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/packet"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/rs/zerolog"
)

const (
	// TODO:
	streamInitTimeout = 10 * time.Second

	// TODO: remove?
	// The timeout for the connection flusher.
	flushTimeout = 7 * time.Second
)

type CallFunc func(ctx *Data) (data interface{}, err error)

// The StreamFunc type describes the function that is called whenever
// a new connection is requested on a peer. It must then handle the new
// connection, if it could be set up correctly.
type StreamFunc func(net.Conn) error

type Session struct {
	closer.Closer

	// The config of this session.
	cf *Config
	// The underlying connection to the remote peer.
	conn Conn
	// The id of the session. The id must be set after the session has been
	// created using SetID().
	id string

	callFuncsMx sync.RWMutex
	callFuncs   map[string]CallFunc
	ctrl        *control

	// Synchronises the access to the stream handlers map.
	streamFuncsMx sync.Mutex
	// The handler functions that have been registered to new streams.
	// When a new stream has been opened, the first data sent on the stream must
	// contain the key into this map to retrieve the correct function to handle
	// the stream.
	streamFuncs map[string]StreamFunc
}

// newSession creates a new orbit session from the given parameters.
// The created session closes, if the underlying connection is closed.
func newSession(cl closer.Closer, conn Conn, initStream net.Conn, cf *Config) (s *Session) {
	s = &Session{
		Closer: cl,
		cf:     cf,
		conn:   conn,
	}
	s.ctrl = newControl(s, initStream)
	s.OnClosing(conn.Close)
	return
}

// ID returns the session ID.
// This must be set manually.
func (s *Session) ID() string {
	return s.id
}

// Todo:
func (s *Session) Codec() codec.Codec {
	return s.cf.Codec
}

func (s *Session) Log() *zerolog.Logger {
	return s.cf.Log
}

// LocalAddr returns the local network address.
func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// Ready signalizes the session that the initialization is done.
// The session starts accepting new incoming streams.
func (s *Session) Ready() {
	go s.acceptStreamRoutine()
}

func (s *Session) RegisterStream(id string, f StreamFunc) {
	s.streamFuncsMx.Lock()
	s.streamFuncs[id] = f
	s.streamFuncsMx.Unlock()
}

// OpenStream opens a new stream with the given channel ID.
// Expires after the timeout and returns a net.Error with Timeout() == true.
// TODO: remove init stream write timeout
func (s *Session) OpenStream(ctx context.Context, id string, t api.StreamType) (stream net.Conn, err error) {
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
	deadline, ok := ctx.Deadline()
	if ok {
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
	if ok {
		err = stream.SetWriteDeadline(time.Time{})
		if err != nil {
			return
		}
	}

	return
}

func (s *Session) RegisterCall(id string, f CallFunc) {
	s.callFuncsMx.Lock()
	s.callFuncs[id] = f
	s.callFuncsMx.Unlock()
}

func (s *Session) Call(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	return s.ctrl.Call(ctx, id, data)
}

func (s *Session) CallAsync(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	stream, err := s.OpenStream(ctx, id, api.StreamTypeCallAsync)
	if err != nil {
		return
	}

	return s.ctrl.CallAsync(ctx, stream, id, data)
}

//###############//
//### Private ###//
//###############//

// acceptStreamRoutine is a routine that accepts new connections and
// calls handleNewStream() for each of them.
// Closes together with the session.
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
				s.cf.Log.Error().
					Err(err).
					Msg("session: failed to accept stream")
			}
			return
		}

		// Run this in a new goroutine.
		go func() {
			err := s.handleNewStream(stream)
			if err != nil {
				s.cf.Log.Error().
					Err(err).
					Msg("session: failed to handle stream")
			}
		}()
	}
}

// handleNewStream handles a new incoming stream. It first reads the initial
// stream data from the connection, which tells us the id of the channel that
// should be opened. With this id, we can retrieve the StreamFunc from the
// session's map and pass it the new stream.
// Recovers from panics.
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

	// Read the initial data from the stream.
	var data api.InitStream
	err = packet.ReadDecode(stream, &data, s.cf.Codec)
	if err != nil {
		return fmt.Errorf("init stream header: %v", err)
	}

	// Decide the type of stream.
	switch data.Type {
	case api.StreamTypeRaw:
		// Obtain the stream handler.
		f, err := s.streamHandler(data.ID)
		if err != nil {
			return
		}

		// Pass it the new stream.
		err = f(stream)
		if err != nil {
			return fmt.Errorf("stream='%v': %v", data.ID, err)
		}
	case api.StreamTypeCallAsync:
		// Pass the stream to the control.
		s.ctrl.HandleCallAsync(stream)
	}

	return
}

// streamHandler retrieves the StreamFunc from the session's map
// that corresponds to the given channel id.
// This method is thread-safe.
func (s *Session) streamHandler(id string) (f StreamFunc, err error) {
	s.streamFuncsMx.Lock()
	f = s.streamFuncs[id]
	s.streamFuncsMx.Unlock()

	if f == nil {
		err = fmt.Errorf("stream handler for id '%s' does not exist", id)
	}
	return
}
