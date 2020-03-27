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

package service

import (
	"fmt"

	"github.com/desertbit/orbit/internal/utils"
	"github.com/desertbit/orbit/pkg/transport"
)

func (s *service) startAcceptConnRoutines() {
	for i := 0; i < s.opts.AcceptConnWorkers; i++ {
		go s.handleNewConnRoutine()
	}
}

func (s *service) handleNewConnRoutine() {
	var (
		err         error
		closingChan = s.ClosingChan()
	)

	for {
		select {
		case <-closingChan:
			return

		case conn := <-s.newConnChan:
			err = s.handleNewConn(conn)
			if err != nil {
				s.log.Error().
					Err(err).
					Msg("service: handle new conn")
			}
		}
	}
}

func (s *service) handleNewConn(conn transport.Conn) (err error) {
	// Generate an id for the session.
	id, err := utils.RandomString(s.opts.SessionIDLen)
	if err != nil {
		return
	}

	// Create a new server session.
	sn, err := initSession(conn, id, s, s.opts)
	if err != nil {
		return
	}

	// Save the session in the map.
	// If the id already exists, close the new session instead.
	// This will happen almost never, if session id len is large enough.
	var idExists bool
	s.sessionsMx.Lock()
	_, idExists = s.sessions[sn.id]
	if !idExists {
		s.sessions[sn.id] = sn
	}
	s.sessionsMx.Unlock()

	// Close the new session, if its id is already taken.
	if idExists {
		sn.Close_()
		return fmt.Errorf("closed new session with duplicate session ID: %s", id)
	}

	// TODO: use a hooker wrapper.
	// Call the hooks.
	/*for _, h := range s.hooks {
		h.OnNewSession(sn)
	}*/

	// Remove the session from the session map, once it closes.
	sn.OnClosing(func() error {
		// Speed up the closing process if the server closes.
		if s.IsClosing() {
			return nil
		}

		s.sessionsMx.Lock()
		delete(s.sessions, sn.id)
		s.sessionsMx.Unlock()
		return nil
	})
	return
}
