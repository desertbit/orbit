/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers@desertbit.com>
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

// ClientSession is used to initialize a new client-side session.
func ClientSession(conn net.Conn, config *Config) (s *Session, err error) {
	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	config = prepareConfig(config)

	// Set a write timeout.
	err = conn.SetWriteDeadline(time.Now().Add(7 * time.Second))
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
	if config.AuthFunc != nil {
		// Reset the deadlines.
		err = conn.SetDeadline(time.Time{})
		if err != nil {
			return
		}

		err = config.AuthFunc(conn)
		if err != nil {
			return
		}
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

	ys, err := yamux.Client(conn, ysConfig)
	if err != nil {
		return
	}

	s = newSession(conn, ys, config, true)
	return
}
