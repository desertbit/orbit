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
	"net"
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/rs/zerolog"
)

type AuthzFunc func(s *Session, id string) (ok bool)

type CallFunc func(ctx context.Context, s *Session, args *Data) (ret interface{}, err error)

type StreamFunc func(s *Session, stream net.Conn) error

type Session struct {
	closer.Closer

	// Arbitrary data that can be set during authentication.
	Value interface{}

	cf    *Config
	log   *zerolog.Logger
	codec codec.Codec
	authz AuthzFunc

	id   string
	conn Conn

	// When a new stream has been opened, the first data sent on the stream must
	// contain the key into this map to retrieve the correct function to handle
	// the stream.
	streamFuncsMx sync.RWMutex
	streamFuncs   map[string]StreamFunc

	callStreamsMx sync.RWMutex
	callStreams   map[string]*callStream // Key: service id

	callFuncsMx sync.RWMutex
	callFuncs   map[string]CallFunc

	callActiveCtxsMx sync.Mutex
	callActiveCtxs   map[uint32]*callContext
}

// newSession creates a new orbit session from the given parameters.
// The created session closes, if the underlying connection is closed.
func newSession(cl closer.Closer, conn Conn, cf *Config) (s *Session) {
	s = &Session{
		Closer:         cl,
		cf:             cf,
		authz:          cf.AuthzFunc,
		log:            cf.Log,
		codec:          cf.Codec,
		conn:           conn,
		streamFuncs:    make(map[string]StreamFunc),
		callStreams:    make(map[string]*callStream),
		callFuncs:      make(map[string]CallFunc),
		callActiveCtxs: make(map[uint32]*callContext),
	}
	s.OnClosing(conn.Close)

	// Start accepting streams.
	go s.acceptStreamRoutine()

	return
}

// ID returns the session ID.
func (s *Session) ID() string {
	return s.id
}

// Codec returns the used transmission codec.
func (s *Session) Codec() codec.Codec {
	return s.codec
}

// LocalAddr returns the local network address.
func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// StreamChanSize returns the size for stream channels.
func (s *Session) StreamChanSize() uint {
	return s.cf.StreamChanSize
}
