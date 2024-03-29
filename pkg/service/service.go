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
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

const (
	NoMaxSizeLimit = -1
	DefaultMaxSize = -2

	DefaultTimeout = 0
	NoTimeout      = -1
)

type (
	CallFunc          func(ctx Context, arg []byte) (ret interface{}, err error)
	RawStreamFunc     func(ctx Context, stream transport.Stream)
	TypedRStreamFunc  func(ctx Context, stream TypedRStream) error
	TypedWStreamFunc  func(ctx Context, stream TypedWStream) error
	TypedRWStreamFunc func(ctx Context, stream TypedRWStream) error
)

type (
	call struct {
		f       CallFunc
		timeout time.Duration
	}

	asyncCallOptions struct {
		maxArgSize int
		maxRetSize int
	}

	streamType uint8

	stream struct {
		typ        streamType
		f          interface{} // One of the above stream funcs.
		maxArgSize int
		maxRetSize int
	}
)

const (
	streamTypeRaw streamType = iota
	streamTypeTR
	streamTypeTW
	streamTypeTRW
)

type Service interface {
	closer.Closer

	// RegisterCall registers a synchronous call send over the shared stream.
	// Do not call after Run() was called.
	// Set timeout to DefaultTimeout for the default timeout.
	// Set timeout to NoTimeout for not timeout.
	RegisterCall(id string, f CallFunc, timeout time.Duration)

	// RegisterAsyncCall registers an asynchronous call using its own stream.
	// Do not call after Run() was called.
	// Set timeout to DefaultTimeout for the default timeout.
	// Set timeout to NoTimeout for not timeout.
	// If maxArgSize & maxRetSize are set to 0, then the payload must be empty.
	// If maxArgSize & maxRetSize are set to NoMaxSizeLimit, then no limit is set.
	// If maxArgSize & maxRetSize are set to DefaultMaxSize, then the default size is used from the options.
	RegisterAsyncCall(id string, f CallFunc, timeout time.Duration, maxArgSize, maxRetSize int)

	// RegisterStream registers the callback for the incoming stream specified by the id.
	// Do not call after Run() was called.
	RegisterStream(id string, f RawStreamFunc)

	// RegisterTypedRStream registers the callback for the incoming typed read stream
	// specified by the id.
	// Do not call after Run() was called.
	// See RegisterAsyncCall() for the usage of maxArgSize.
	RegisterTypedRStream(id string, f TypedRStreamFunc, maxArgSize int)

	// RegisterTypedWStream registers the callback for the incoming typed write stream
	// specified by the id.
	// Do not call after Run() was called.
	// See RegisterAsyncCall() for the usage of maxRetSize.
	RegisterTypedWStream(id string, f TypedWStreamFunc, maxRetSize int)

	// RegisterTypedRWStream registers the callback for the incoming typed read write stream
	// specified by the id.
	// Do not call after Run() was called.
	// See RegisterAsyncCall() for the usage of maxArgSize & maxRetSize.
	RegisterTypedRWStream(id string, f TypedRWStreamFunc, maxArgSize, maxRetSize int)

	// Run the service and start accepting requests.
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

	streams       map[string]stream           // Key: streamID
	calls         map[string]call             // Key: callID
	asyncCallOpts map[string]asyncCallOptions // Key: callID
}

func New(opts *Options) (Service, error) {
	opts.setDefaults()
	err := opts.validate()
	if err != nil {
		return nil, err
	}

	s := &service{
		Closer:        opts.Closer,
		opts:          opts,
		codec:         opts.Codec,
		log:           opts.Log,
		hooks:         opts.Hooks,
		newConnChan:   make(chan transport.Conn, opts.AcceptConnWorkers),
		sessions:      make(map[string]*session),
		streams:       make(map[string]stream),
		calls:         make(map[string]call),
		asyncCallOpts: make(map[string]asyncCallOptions),
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

func (s *service) RegisterCall(id string, f CallFunc, timeout time.Duration) {
	// Use default options if required.
	if timeout == DefaultTimeout {
		timeout = s.opts.CallTimeout
	}

	// Save the call.
	s.calls[id] = call{
		f:       f,
		timeout: timeout,
	}
}

func (s *service) RegisterAsyncCall(id string, f CallFunc, timeout time.Duration, maxArgSize, maxRetSize int) {
	s.RegisterCall(id, f, timeout)

	// Use default options if required.
	if maxArgSize == DefaultMaxSize {
		maxArgSize = s.opts.MaxArgSize
	}
	if maxRetSize == DefaultMaxSize {
		maxRetSize = s.opts.MaxRetSize
	}

	// Save the limits for later.
	s.asyncCallOpts[id] = asyncCallOptions{
		maxArgSize: maxArgSize,
		maxRetSize: maxRetSize,
	}
}

func (s *service) RegisterStream(id string, f RawStreamFunc) {
	s.streams[id] = stream{typ: streamTypeRaw, f: f}
}

func (s *service) RegisterTypedRStream(id string, f TypedRStreamFunc, maxArgSize int) {
	// Use default options if required.
	if maxArgSize == DefaultMaxSize {
		maxArgSize = s.opts.MaxArgSize
	}

	s.streams[id] = stream{
		typ:        streamTypeTR,
		f:          f,
		maxArgSize: maxArgSize,
	}
}

func (s *service) RegisterTypedWStream(id string, f TypedWStreamFunc, maxRetSize int) {
	// Use default options if required.
	if maxRetSize == DefaultMaxSize {
		maxRetSize = s.opts.MaxRetSize
	}

	s.streams[id] = stream{
		typ:        streamTypeTW,
		f:          f,
		maxRetSize: maxRetSize,
	}
}

func (s *service) RegisterTypedRWStream(id string, f TypedRWStreamFunc, maxArgSize, maxRetSize int) {
	// Use default options if required.
	if maxArgSize == DefaultMaxSize {
		maxArgSize = s.opts.MaxArgSize
	}
	if maxRetSize == DefaultMaxSize {
		maxRetSize = s.opts.MaxRetSize
	}

	s.streams[id] = stream{
		typ:        streamTypeTRW,
		f:          f,
		maxArgSize: maxArgSize,
		maxRetSize: maxRetSize,
	}
}
