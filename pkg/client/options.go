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

package client

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

type Options struct {
	// Transport specifies the communication backend. This value must be set.
	Transport transport.Transport

	// Optional values:
	// ################

	// TransportValue can be used to pass arbitrary arguments to the transport's Dial().
	// If supported by the transport, it should provide a Value() method to construct a correct value.
	TransportValue interface{}

	// Closer defines the closer instance. A default closer will be created if unspecified.
	Closer closer.Closer

	// Codec defines the transport encoding. A default codec will be used if unspecified.
	Codec codec.Codec

	// Hooks specifies the hooks executed during certain actions. The order of the hooks is stricly followed.
	Hooks Hooks

	// Log specifies the default logger backend. A default logger will be used if unspecified.
	Log *zerolog.Logger

	// CallTimeout specifies the default timeout for each call.
	// Set to -1 for no timeout.
	CallTimeout time.Duration

	// ConnectTimeout specifies the timeout duration after a service connect attempt.
	ConnectTimeout time.Duration

	// ConnectThrottleDuration specifies the wait duration between subsequent connection attempts.
	ConnectThrottleDuration time.Duration

	// HandshakeTimeout specifies the connection initialization timeout.
	HandshakeTimeout time.Duration

	// StreamInitTimeout specifies the default timeout for a stream setup.
	StreamInitTimeout time.Duration

	// PrintPanicStackTraces prints stack traces of catched panics.
	PrintPanicStackTraces bool

	// MaxArgSize defines the default maximum argument payload size for RPC calls.
	MaxArgSize int

	// MaxRetSize defines the default maximum return payload size for RPC calls.
	MaxRetSize int

	// MaxHeaderSize defines the maximum header size for calls and streams.
	MaxHeaderSize int
}

func DefaultOptions(tr transport.Transport) Options {
	return Options{
		Transport:               tr,
		Closer:                  closer.New(),
		Codec:                   msgpack.Codec,
		Log:                     defaultLogger(),
		CallTimeout:             30 * time.Second,
		ConnectTimeout:          10 * time.Second,
		ConnectThrottleDuration: 2 * time.Second,
		HandshakeTimeout:        7 * time.Second,
		StreamInitTimeout:       10 * time.Second,
		MaxArgSize:              4 * 1024 * 1024, // 4 MB
		MaxRetSize:              4 * 1024 * 1024, // 4 MB
		MaxHeaderSize:           500 * 1024,      // 500 KB
	}
}

func defaultLogger() *zerolog.Logger {
	l := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Str("component", "orbit").Logger()
	return &l
}

func (o *Options) validate() error {
	if o.Transport == nil {
		return errors.New("no transport set")
	}
	return nil
}
