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
	"errors"
	"os"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

const (
	defaultAcceptConnWorkers = 5
	defaultSessionIDLen      = 32
	defaultHandshakeTimeout  = 7 * time.Second
)

type Options struct {
	// ListenAddr specifies the listen address for the server. This value is passed to the transport backend.
	ListenAddr string

	// Transport specifies the communication backend. This value must be set.
	Transport transport.Transport

	// Optional values:
	// ################

	// Closer defines the closer instance. A default closer will be created if unspecified.
	Closer closer.Closer

	// Codec defines the transport encoding. A default codec will be used if unspecified.
	Codec codec.Codec

	// Hooks specifies the hooks executed during certain actions. The order of the hooks is stricly followed.
	Hooks Hooks

	// Log specifies the default logger backend. A default logger will be used if unspecified.
	Log *zerolog.Logger

	// AcceptConnWorkers specifies the routines accepting new connections.
	AcceptConnWorkers int

	// SessionIDLen specifies the length of each session ID.
	SessionIDLen uint

	// HandshakeTimeout specifies the timeout for the initial handshake.
	HandshakeTimeout time.Duration

	// PrintPanicStackTraces prints stack traces of catched panics.
	PrintPanicStackTraces bool

	// SendInternalErrors sends all errors to the client, even if the Error interface is not satisfied.
	SendInternalErrors bool
}

func (o *Options) setDefaults() {
	if o.Closer == nil {
		o.Closer = closer.New()
	}
	if o.Codec == nil {
		o.Codec = msgpack.Codec
	}
	if o.Log == nil {
		l := zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Str("component", "orbit").Logger()
		o.Log = &l
	}
	if o.AcceptConnWorkers == 0 {
		o.AcceptConnWorkers = defaultAcceptConnWorkers
	}
	if o.SessionIDLen == 0 {
		o.SessionIDLen = defaultSessionIDLen
	}
	if o.HandshakeTimeout == 0 {
		o.HandshakeTimeout = defaultHandshakeTimeout
	}
}

func (o *Options) validate() error {
	if o.ListenAddr == "" {
		return errors.New("empty listen address")
	} else if o.Transport == nil {
		return errors.New("no transport set")
	}
	return nil
}
