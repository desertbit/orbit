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

package client

import (
	"context"
	"time"

	"github.com/desertbit/orbit/internal/throttler"
)

func (c *client) getSession() (s *session) {
	c.sessionMx.Lock()
	s = c.session
	c.sessionMx.Unlock()
	return
}

func (c *client) setSession(s *session) {
	c.sessionMx.Lock()
	c.session = s
	c.sessionMx.Unlock()
}

// connectedSession returns the connected session or triggers a connect request.
// Fails if no connection could be established.
// Always returns a connected session if err is nil.
func (c *client) connectedSession(ctx context.Context) (s *session, err error) {
	s = c.getSession()
	if s != nil {
		return
	}

	var (
		closingChan = c.ClosingChan()
		retChan     = make(chan *session, 1)
		ctxDone     = ctx.Done()
	)

	select {
	case <-closingChan:
		err = ErrClosed
		return
	case <-ctxDone:
		err = ctx.Err()
		return
	case c.connectSessionChan <- retChan:
	}

	select {
	case <-closingChan:
		err = ErrClosed
		return
	case <-ctxDone:
		err = ctx.Err()
		return
	case s = <-retChan:
		if s == nil {
			err = ErrNoSession
		}
		return
	}
}

func (c *client) startSessionRoutine() {
	go c.sessionRoutine()
}

func (c *client) sessionRoutine() {
	defer c.Close_()

	var (
		closingChan      = c.ClosingChan()
		connectThrottler = throttler.New(c.opts.ConnectThrottleDuration)
	)

Loop:
	for {
		select {
		case <-closingChan:
			return

		case r := <-c.connectSessionChan:
			// Throttle between connection attempts.
			connectThrottler.ThrottleSleep(time.Now())

			// Try to connect to session.
			s, err := connectSession(c, c.opts)
			if err != nil {
				c.log.Error().Err(err).Msg("failed to connect client session")
				r <- nil // Notify.
				continue Loop
			}

			// Publish the newly connected session.
			c.setSession(s)

			// Notify.
			r <- s

			// Wait for disconnection.
			// Just handle the connectSessionChan and keep it drained.
			sessionClosingChan := s.ClosingChan()
		SubLoop:
			for {
				select {
				case <-closingChan:
					s.Close()
					return
				case <-sessionClosingChan:
					break SubLoop
				case r := <-c.connectSessionChan:
					r <- s
				}
			}

			// Reset the session again.
			c.setSession(nil)
		}
	}
}
