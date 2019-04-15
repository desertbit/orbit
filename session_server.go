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
	"fmt"
	"net"
	"time"

	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/internal/api"
	"github.com/hashicorp/yamux"
)

const (
	// The time duration after which we timeout if the version byte could
	// not be written to the stream.
	streamVersionReadTimeout = 20 * time.Second
)

// ServerSession is used to initialize a new server-side session.
// A config can be provided, where every property of it that has not
// been set will be initialized with a default value.
// That makes it possible to overwrite only the interesting properties
// for the caller.
// When the connection has been established, the server waits for the
// API version byte of the client. If the versions do not match, the
// server immediately closes the connection.
// As part of the session setup, the auth func of the config is called,
// if it has been defined.
func ServerSession(conn net.Conn, config *Config) (s *Session, err error) {
	return serverSession(conn, config, closer.New())
}

// ServerSessionWithCloser initializes a new server-side session, just
// as ServerSession() does, but allows to hand in an own closer.
func ServerSessionWithCloser(conn net.Conn, config *Config, cl closer.Closer) (s *Session, err error) {
	return serverSession(conn, config, cl)
}

// serverSession is the internal helper to initialize a new server-side session.
func serverSession(conn net.Conn, config *Config, cl closer.Closer) (s *Session, err error) {
	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// Prepare the config with default values, where needed.
	config = prepareConfig(config)

	// Set a read timeout.
	err = conn.SetReadDeadline(time.Now().Add(streamVersionReadTimeout))
	if err != nil {
		return
	}

	// Ensure client has the same protocol version.
	v := make([]byte, 1)
	n, err := conn.Read(v)
	if err != nil {
		return
	} else if n != 1 {
		return nil, fmt.Errorf("failed to read version byte from connection")
	} else if v[0] != api.Version {
		return nil, ErrInvalidVersion
	}

	// Authenticate if required.
	value, err := authSession(conn, config)
	if err != nil {
		return
	}

	// Reset the deadlines.
	err = conn.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Prepare yamux config.
	ysConfig := yamux.DefaultConfig()
	ysConfig.Logger = config.Logger
	ysConfig.LogOutput = nil
	ysConfig.KeepAliveInterval = config.KeepAliveInterval
	ysConfig.EnableKeepAlive = true

	// Create a new yamux server session.
	ys, err := yamux.Server(conn, ysConfig)
	if err != nil {
		return
	}

	// Finally, create the orbit server session.
	s = newSession(conn, ys, config, false, cl)

	// Save the arbitrary data from the auth func.
	s.Value = value
	return
}
