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

package auth

import (
	"github.com/desertbit/orbit/sample/api"
	"golang.org/x/crypto/bcrypt"
	"net"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/packet"
)

type GetHashHook func(username string) (hash []byte, ok bool)

func Server(hook GetHashHook) orbit.AuthFunc {
	if hook == nil {
		panic("auth hook for server must not be nil")
	}

	return func(conn net.Conn) (value interface{}, err error) {
		// Read in incoming authentication request.
		var data api.AuthRequest
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

		// Validate the request data.
		if data.Username == "" {
			return nil, errInvalidUsername
		} else if len(data.Pw) == 0 {
			return nil, errInvalidPassword
		}

		var ret api.AuthResponse

		// Retrieve the hash for the user.
		pwHash, ok := hook(data.Username)

		// Determine if the sent password is correct.
		ret.Ok = ok && (bcrypt.CompareHashAndPassword(pwHash, data.Pw) == nil)

		// Send a response back to the client if
		// the authentication was successful.
		err = packet.WriteEncode(
			conn,
			&ret,
			msgpack.Codec,
			maxPayloadSize,
			writeTimeout,
		)
		if err != nil {
			return
		}

		// Check if the authentication was successful.
		if !ret.Ok {
			err = errAuthFailed
			return
		}

		// Pass the username to the session value.
		value = data.Username
		return
	}
}
