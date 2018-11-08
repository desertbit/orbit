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

package auth

import (
	"github.com/desertbit/orbit/sample/api"
	"net"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/packet"
)

func Client(username, pw string) orbit.AuthFunc {
	// Generate a checksum of the password.
	checksum := Checksum(pw)
	return func(conn net.Conn) (value interface{}, err error) {
		// Send an authentication request with our credentials.
		err = packet.WriteEncode(
			conn,
			&api.AuthRequest{
				Username: username,
				Pw:       checksum[:],
			},
			msgpack.Codec,
			maxPayloadSize,
			writeTimeout,
		)
		if err != nil {
			return
		}

		// Read the server's response for it.
		var data api.AuthResponse
		err = packet.ReadDecode(
			conn,
			&data,
			msgpack.Codec,
			maxPayloadSize,
			readTimeout,
		)
		if err != nil {
			return
		}

		// Check if the authentication was successful.
		if !data.Ok {
			err = errAuthFailed
			return
		}

		return
	}
}
