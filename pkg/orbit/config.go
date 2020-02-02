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
	"os"
	"time"

	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/rs/zerolog"
)

const (
	// todo:
	defaultInitTimeout = 10 * time.Second
	// todo:
	defaultCallTimeout = 30 * time.Second
)

type Config struct {
	Log   *zerolog.Logger
	Codec codec.Codec

	PrintPanicStackTraces bool
	SendErrToCaller       bool

	// InitTimeout specifies the connection initialization timeout.
	InitTimeout time.Duration

	// CallTimeout specifies the default timeout for any Call.
	// This value can be overridden per call in the .orbit file.
	// A value < 0 means no timeout.
	// A value of 0 will take the default call timeout value.
	CallTimeout time.Duration
}

func prepareConfig(c *Config) *Config {
	if c == nil {
		c = &Config{}
	}

	if c.Log == nil {
		l := zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Str("component", "orbit").Logger()
		c.Log = &l
	}
	if c.Codec == nil {
		c.Codec = msgpack.Codec
	}

	if c.InitTimeout == 0 {
		c.InitTimeout = defaultInitTimeout
	}
	if c.CallTimeout == 0 {
		c.CallTimeout = defaultCallTimeout
	}
	return c
}
