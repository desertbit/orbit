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

	"github.com/desertbit/orbit/sample/api"

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
