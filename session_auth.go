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
	"net"
	"time"

	"github.com/desertbit/orbit/internal/flusher"
)

// authSession calls the authentication func defined in the given config.
// If no such func has been defined, it returns immediately.
// This is a convenience func that ensures the connection is flushed,
// even if the authentication fails. That ensures that errors can be
// handled appropriately on each peer's side.
// Can be called for both the server and client side.
func authSession(conn net.Conn, config *Config) (v interface{}, err error) {
	// Skip if no authentication hook is defined.
	if config.AuthFunc == nil {
		return
	}

	// Always flush the connection on defer.
	// Otherwise authentication errors might not be send
	// to the peer, because the connection is closed too fast.
	defer func() {
		derr := flusher.Flush(conn, flushTimeout)
		if err == nil {
			err = derr
		}
	}()

	// Reset the deadlines.
	err = conn.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Call the auth func defined in the config.
	return config.AuthFunc(conn)
}
