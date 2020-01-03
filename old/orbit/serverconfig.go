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

package orbit

const (
	// The default number of workers accepting new connections.
	defaultNewConnNumberWorkers = 5

	// The default size of the channel newly accepted connections are
	// passed into to be further processed.
	defaultNewConnChanSize = 5

	// The default size of the channel newly created sessions are
	// passend into to be further processed by users of this package.
	defaultNewSessionChanSize = 5
)

type ServerConfig struct {
	// Embed the standard config that both clients and servers share.
	*Config

	// The number of goroutines that handle incoming connections on the server.
	NewConnNumberWorkers int
	// The size of the channel on which new connections are passed to the
	// server workers.
	// Should not be less than NewConnNumberWorkers.
	NewConnChanSize int
	// The size of the channel on which new server sessions are passed into,
	// so that a user of this package can read them from it.
	// Should not be less than NewConnNumberWorkers.
	NewSessionChanSize int
}

// prepareServerConfig assigns default values to each property of the given config,
// if it has not been set. If a nil config is provided, a new one is created.
// The final config is returned.
func prepareServerConfig(c *ServerConfig) *ServerConfig {
	if c == nil {
		c = &ServerConfig{}
	}

	// Prepare the standard config.
	c.Config = prepareConfig(c.Config)

	if c.NewConnNumberWorkers == 0 {
		c.NewConnNumberWorkers = defaultNewConnNumberWorkers
	}
	if c.NewConnChanSize == 0 {
		c.NewConnChanSize = defaultNewConnChanSize
	}
	if c.NewSessionChanSize == 0 {
		c.NewSessionChanSize = defaultNewSessionChanSize
	}
	return c
}
