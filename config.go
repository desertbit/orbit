/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
