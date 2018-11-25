/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
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
	"fmt"
	"net"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/hashicorp/yamux"
)

const (
	// The time duration after which we timeout if the version byte could not
	// be written to the stream.
	streamVersionWriteTimeout = 15 * time.Second
)

// ClientSession is used to initialize a new client-side session.
// A config can be provided, where every property of it that has not
// been set will be initialized with a default value.
// That makes it possible to overwrite only the interesting properties
// for the caller.
// When the connection to the server has been established, one byte
// containing the version of the client is sent to the server.
// In case the versions do not match, the server will close the connection.
// As part of the session setup, the auth func of the config is called,
// if it has been defined.
func ClientSession(conn net.Conn, config *Config) (s *Session, err error) {
	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// Prepare the config with default values, where needed.
	config = prepareConfig(config)

	// Set a write timeout.
	err = conn.SetWriteDeadline(time.Now().Add(streamVersionWriteTimeout))
	if err != nil {
		return
	}

	// Tell the server our protocol version.
	v := []byte{api.Version}
	n, err := conn.Write(v)
	if err != nil {
		return
	} else if n != 1 {
		return nil, fmt.Errorf("failed to write version byte to connection")
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

	// Create a new yamux client session.
	ys, err := yamux.Client(conn, ysConfig)
	if err != nil {
		return
	}

	// Finally, create the orbit client session.
	s = newSession(conn, ys, config, true)

	// Save the arbitrary data from the auth func.
	s.Value = value
	return
}
