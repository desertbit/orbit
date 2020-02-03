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
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/rs/zerolog"
)

type CallFunc func(ctx context.Context, s *Session, args *Data) (ret interface{}, err error)

type StreamFunc func(s *Session, stream net.Conn)

type Session struct {
	closer.Closer

	cf    *Config
	log   *zerolog.Logger
	codec codec.Codec
	hooks []Hook

	id   string
	conn Conn

	valuesMx sync.RWMutex
	values   map[string]interface{}

	// When a new stream has been opened, the first data sent on the stream must
	// contain the key into this map to retrieve the correct function to handle
	// the stream.
	streamFuncsMx sync.RWMutex
	streamFuncs   map[string]StreamFunc

	callStreamsMx sync.RWMutex
	callStreams   map[string]*callStream // Key: service id

	callFuncsMx sync.RWMutex
	callFuncs   map[string]CallFunc // Key: serviceID.callID

	callActiveCtxsMx sync.Mutex
	callActiveCtxs   map[uint32]*callContext
}

// newSession creates a new orbit session from the given parameters.
// The created session closes, if the underlying connection is closed.
// The initial stream is closed by the caller.
func newSession(
	cl closer.Closer,
	conn Conn,
	initStream net.Conn,
	id string,
	cf *Config,
	hs []Hook,
) (s *Session, err error) {
	s = &Session{
		Closer:         cl,
		cf:             cf,
		log:            cf.Log,
		codec:          cf.Codec,
		hooks:          hs,
		id:             id,
		conn:           conn,
		values:         make(map[string]interface{}),
		streamFuncs:    make(map[string]StreamFunc),
		callStreams:    make(map[string]*callStream),
		callFuncs:      make(map[string]CallFunc),
		callActiveCtxs: make(map[uint32]*callContext),
	}
	s.OnClosing(conn.Close)

	// Catch panics.
	defer func() {
		if e := recover(); e != nil {
			if cf.PrintPanicStackTraces {
				err = fmt.Errorf("catched panic: \n%v\n%s", e, string(debug.Stack()))
			} else {
				err = fmt.Errorf("catched panic: \n%v", e)
			}
		}
	}()

	// Call OnNewSession hooks.
	for _, h := range hs {
		err = h.OnNewSession(s, initStream)
		if err != nil {
			return
		}
	}

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

// CallTimeout returns the default timeout for all calls.
func (s *Session) CallTimeout() time.Duration {
	return s.cf.CallTimeout
}

// SetValue saves the value v for the key k in the values map.
func (s *Session) SetValue(k string, v interface{}) {
	s.valuesMx.Lock()
	s.values[k] = v
	s.valuesMx.Unlock()
}

// Value returns the value for the key k from the values map.
// If the key does not exist, v == nil.
func (s *Session) Value(k string) (v interface{}) {
	s.valuesMx.RLock()
	v = s.values[k]
	s.valuesMx.RUnlock()
	return
}

// DeleteValue deletes the value for the key k in the values map.
// If the key does not exist, this is a no-op.
func (s *Session) DeleteValue(k string) {
	s.valuesMx.Lock()
	delete(s.values, k)
	s.valuesMx.Unlock()
}
