/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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

/*
 *  Juno
 *  Copyright (C) 2016 DesertBit
 */

package auth

import (
	"fmt"
	"time"

	"github.com/desertbit/orbit/pkg/hook/auth/api"
	"github.com/desertbit/orbit/pkg/hook/auth/bcrypt"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/packet"
)

type UserHashFunc func(username string) (hash []byte, ok bool)

var _ orbit.Hook = &server{}

type server struct {
	hf UserHashFunc
}

func ServerHook(hf UserHashFunc) orbit.Hook {
	if hf == nil {
		panic("user hash func for server must not be nil")
	}

	return &server{hf: hf}
}

// Implements the orbit.Hook interface.
func (s *server) OnNewSession(sn *orbit.Session, stream orbit.Stream) (err error) {
	cc := sn.Codec()

	// Set a deadline.
	err = stream.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return fmt.Errorf("auth set deadline: %w", err)
	}

	// Read an incoming authentication request.
	var data api.Request
	err = packet.ReadDecode(stream, &data, cc)
	if err != nil {
		return fmt.Errorf("auth packet read decode: %w", err)
	}

	// Validate the request data.
	if data.Username == "" {
		return errInvalidUsername
	} else if len(data.Pw) == 0 {
		return errInvalidPassword
	}

	// Retrieve the hash for the user.
	pwHash, ok := s.hf(data.Username)

	// Determine if the sent password is correct.
	ret := api.Response{
		Ok: ok && (bcrypt.Compare(pwHash, data.Pw) == nil),
	}

	// Send a response back to the client if
	// the authentication was successful.
	err = packet.WriteEncode(stream, &ret, cc)
	if err != nil {
		return fmt.Errorf("auth packet write encode: %w", err)
	}

	// Check if the authentication was successful.
	if !ret.Ok {
		err = errAuthFailed
		return
	}

	// Save the username to the session values.
	sn.SetValue(keyUsername, data.Username)

	return
}

// Implements the orbit.Hook interface.
func (s *server) OnCall(sn *orbit.Session, service, id string) error { return nil }

// Implements the orbit.Hook interface.
func (s *server) OnCallCompleted(sn *orbit.Session, service, id string, err error) {}

// Implements the orbit.Hook interface.
func (s *server) OnNewStream(sn *orbit.Session, service, id string) error { return nil }
