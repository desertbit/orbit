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
	"net"
	"time"

	"github.com/desertbit/orbit/internal/flusher"
	"github.com/desertbit/orbit/pkg/hook/auth/api"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/packet"
)

var _ orbit.Hook

type client struct {
	username   string
	pwChecksum []byte
}

func ClientHook(username, pw string) orbit.Hook {
	cs := Checksum(pw)
	return &client{
		username:   username,
		pwChecksum: cs[:],
	}
}

// Implements the orbit.Hook interface.
func (c *client) OnNewSession(s *orbit.Session, stream net.Conn) (err error) {
	cc := s.Codec()

	// Set a deadline.
	err = stream.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	// Send an authentication request with our credentials.
	err = packet.WriteEncode(stream, &api.Request{Username: c.username, Pw: c.pwChecksum}, cc)
	if err != nil {
		return
	}

	// Read the server's response for it.
	var data api.Response
	err = packet.ReadDecode(stream, &data, cc)
	if err != nil {
		return
	}

	// Always flush the connection.
	// Otherwise authentication errors might not be send
	// to the peer, because the connection is closed too fast.
	err = flusher.Flush(stream, flushTimeout)
	if err != nil {
		return
	}

	// Check if the authentication was successful.
	if !data.Ok {
		return errAuthFailed
	}

	return
}

// Implements the orbit.Hook interface.
func (*client) OnNewCall(s *orbit.Session, service, id string) error { return nil }

// Implements the orbit.Hook interface.
func (*client) OnCallCompleted(s *orbit.Session, service, id string, err error) {}

// Implements the orbit.Hook interface.
func (*client) OnNewStream(s *orbit.Session, service, id string) error { return nil }
