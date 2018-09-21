/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/codec/msgpack"
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
}

func (c *Config) SetDefaults() {
	if c.Codec == nil {
		c.Codec = msgpack.Codec
	}
	if c.KeepAliveInterval == 0 {
		c.KeepAliveInterval = 30 * time.Second
	}
	if c.Logger == nil {
		c.Logger = log.New(os.Stderr, "orbit: ", 0)
	}
}

func prepareConfig(c *Config) *Config {
	if c == nil {
		c = &Config{}
	}
	c.SetDefaults()
	return c
}
