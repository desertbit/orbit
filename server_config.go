/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
