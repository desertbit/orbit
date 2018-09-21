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
package control

import (
	"log"
	"os"
	"time"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/codec/msgpack"
)

const (
	// defaultMaxMessageSize specifies the default maximum message payload size in KiloBytes.
	defaultMaxMessageSize = 100 * 1024 // 100KB

	// defaultCallTimeout specifies the default timeout for a call request.
	defaultCallTimeout = 30 * time.Second

	// defaultReadTimeout specifies ...
	defaultReadTimeout = 35 * time.Second

	// defaultWriteTimeout
	defaultWriteTimeout = 15 * time.Second
)

var (
	defaultCodec  = msgpack.Codec
	defaultLogger = log.New(os.Stderr, "orbit: control: ", 0)
)

type Config struct {
	Codec          codec.Codec
	Logger         *log.Logger
	MaxMessageSize int
	CallTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

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
