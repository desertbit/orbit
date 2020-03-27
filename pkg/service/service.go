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
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

type (
	CallFunc   func(ctx Context, arg []byte) (ret interface{}, err error)
	StreamFunc func(ctx Context, stream transport.Stream)
)

type Service interface {
	closer.Closer

	RegisterCall(id string, f CallFunc)
	RegisterStream(id string, f StreamFunc)
	Run() error
}

type service struct {
	closer.Closer

	opts  *Options
	codec codec.Codec
	log   *zerolog.Logger
	hooks Hooks

	newConnChan chan transport.Conn

	sessionsMx sync.RWMutex
	sessions   map[string]*session

	// When a new stream has been opened, the first data sent on the stream must
	// contain the key into this map to retrieve the correct function to handle
	// the stream.
	streamFuncsMx sync.RWMutex
	streamFuncs   map[string]StreamFunc // Key: streamID

	callFuncsMx sync.RWMutex
	callFuncs   map[string]CallFunc // Key: callID
}

func New(opts *Options) (Service, error) {
	opts.setDefaults()
	err := opts.validate()
	if err != nil {
		return nil, err
	}

	s := &service{
		Closer:      opts.Closer,
		opts:        opts,
		codec:       opts.Codec,
		log:         opts.Log,
		hooks:       opts.Hooks,
		newConnChan: make(chan transport.Conn, opts.AcceptConnWorkers),
		sessions:    make(map[string]*session),
		streamFuncs: make(map[string]StreamFunc),
		callFuncs:   make(map[string]CallFunc),
	}
	s.OnClose(s.hookClose)
	s.startAcceptConnRoutines()
	return s, nil
}

// Run the service and start listening for requests.
// This method is blocking.
func (s *service) Run() (err error) {
	defer s.Close_()

	// Open a listener with the transport.
	ln, err := s.opts.Transport.Listen(s.CloserTwoWay(), s.opts.ListenAddr)
	if err != nil {
		return
	}

	s.log.Info().Str("listenAddr", s.opts.ListenAddr).Msg("service listening")

	var (
		conn        transport.Conn
		closingChan = s.ClosingChan()
	)

	for {
		// Listen for incoming connections.
		conn, err = ln.Accept()
		if err != nil {
			if s.IsClosing() {
				err = nil
				return
			}
			return
		}

		// Pass the conn to the handle routines.
		select {
		case <-closingChan:
			return
		case s.newConnChan <- conn:
		}
	}
}

func (s *service) RegisterCall(id string, f CallFunc) {
	s.callFuncsMx.Lock()
	s.callFuncs[id] = f
	s.callFuncsMx.Unlock()
}

func (s *service) callFunc(id string) (f CallFunc, ok bool) {
	s.callFuncsMx.RLock()
	f, ok = s.callFuncs[id]
	s.callFuncsMx.RUnlock()
	return
}

func (s *service) RegisterStream(id string, f StreamFunc) {
	s.streamFuncsMx.Lock()
	s.streamFuncs[id] = f
	s.streamFuncsMx.Unlock()
}

func (s *service) streamFunc(id string) (f StreamFunc, ok bool) {
	s.streamFuncsMx.RLock()
	f, ok = s.streamFuncs[id]
	s.streamFuncsMx.RUnlock()
	return
}
