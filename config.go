/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
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
	"log"
	"os"
	"time"

	"github.com/desertbit/orbit/codec/msgpack"

	"github.com/desertbit/orbit/codec"
)

const (
	// defaultKeepAliveInterval is the default keep alive interval that
	// is passed to the yamux sessions.
	defaultKeepAliveInterval = 30 * time.Second
)

var (
	// defaultCodec is the default codec being used to encode/decode messages in orbit.
	// Defaults to msgpack.
	defaultCodec = msgpack.Codec

	// defaultLogger that is used to log messages to.
	// Defaults to os.Stderr.
	defaultLogger = log.New(os.Stderr, "orbit: ", 0)
)

type Config struct {
	// Codec used to encode and decode orbit messages.
	// Defaults to msgpack.
	Codec codec.Codec

	// KeepAliveInterval is how often to perform the keep alive.
	// Default to 30 seconds.
	KeepAliveInterval time.Duration

	// Logger is used to pass in the logger to be used.
	// Uses a default logger to os.Stderr.
	Logger *log.Logger

	// AuthFunc authenticates the session connection if defined.
	// It gets called right after the version byte has been exchanged
	// between client and server. Therefore, not much resources are wasted
	// in case the authentication fails.
	AuthFunc AuthFunc
}

// prepareConfig assigns default values to each property of the given config,
// if it has not been set. If a nil config is provided, a new one is created.
// The final config is returned.
func prepareConfig(c *Config) *Config {
	if c == nil {
		c = &Config{}
	}
	if c.Codec == nil {
		c.Codec = defaultCodec
	}
	if c.KeepAliveInterval == 0 {
		c.KeepAliveInterval = defaultKeepAliveInterval
	}
	if c.Logger == nil {
		c.Logger = defaultLogger
	}
	return c
}
