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

package control

import (
	"log"
	"os"
	"time"

	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
)

const (
	// defaultMaxMessageSize specifies the default maximum message payload size in KiloBytes.
	defaultMaxMessageSize = 100 * 1024 // 100KB

	// defaultCallTimeout specifies the default timeout for a call request.
	defaultCallTimeout = 30 * time.Second

	// defaultReadTimeout specifies the default timeout for reading from the connection.
	defaultReadTimeout = 35 * time.Second

	// defaultWriteTimeout specifies the default timeout for writing to the connection.
	defaultWriteTimeout = 15 * time.Second
)

var (
	// defaultCodec is the codec that is used by default, message pack currently.
	defaultCodec = msgpack.Codec

	// defaultLogger is the logger that is used by default.
	defaultLogger = log.New(os.Stderr, "orbit: ", 0)
)

// The Config type contains the possible configuration parameter of a Control.
type Config struct {
	// The codec.Codec that should be used to encode the payload of messages.
	Codec codec.Codec

	// The log.logger to be used for writing log messages to.
	Logger *log.Logger

	// The max size in bytes a message may have that is sent over the stream.
	MaxMessageSize int

	// The maximum time a call may take to finish.
	CallTimeout time.Duration
	// The maximum time it may take to read a packet from the stream.
	ReadTimeout time.Duration
	// The maximum time it may take to write a packet to the stream.
	WriteTimeout time.Duration

	// A flag that controls whether we sent errors that occur during the
	// handling of calls back to the caller.
	// Be cautious, in case you set this to true, as errors may contain
	// critical information about your app, most often useful for attackers.
	SendErrToCaller bool
}

// prepareConfig sets default values on the properties of
// the given config, that are not set yet.
// If a nil Config is provided, a new config is created
// that consequently contains only default values.
func prepareConfig(c *Config) *Config {
	if c == nil {
		c = &Config{}
	}
	if c.Codec == nil {
		c.Codec = defaultCodec
	}
	if c.Logger == nil {
		c.Logger = defaultLogger
	}
	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = defaultMaxMessageSize
	}
	if c.CallTimeout == 0 {
		c.CallTimeout = defaultCallTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	return c
}
