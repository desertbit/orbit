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

package yamux

import (
	"crypto/tls"
	"errors"
	"os"
	"time"

	"github.com/desertbit/yamux"
	"github.com/rs/zerolog"
)

type Options struct {
	// TODO:
	ListenAddr string

	// TODO:
	DialAddr string

	// TODO:
	Config *yamux.Config

	// TODO:
	// If nil, no encryption.
	TLSConfig *tls.Config
}

func DefaultOptions(listenAddr, dialAddr string, tlsc *tls.Config) Options {
	return Options{
		ListenAddr: listenAddr,
		DialAddr:   dialAddr,
		Config: &yamux.Config{
			AcceptBacklog:          256,
			EnableKeepAlive:        true,
			KeepAliveInterval:      30 * time.Second,
			ConnectionWriteTimeout: 10 * time.Second,
			MaxStreamWindowSize:    256 * 1024,
			LogOutput:              defaultLogger(),
		},
		TLSConfig: tlsc,
	}
}

func (o Options) validate() (err error) {
	if o.ListenAddr == "" {
		return errors.New("listen address not set")
	}
	if o.DialAddr == "" {
		return errors.New("dial address not set")
	}
	err = yamux.VerifyConfig(o.Config)
	if err != nil {
		return
	}
	return
}

func defaultLogger() *zerolog.Logger {
	l := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Str("component", "orbit.yamux").Logger()
	return &l
}
