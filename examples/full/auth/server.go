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

package auth

import (
	"net"

	"github.com/desertbit/orbit/examples/full/api"
	"golang.org/x/crypto/bcrypt"

	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/packet"
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
