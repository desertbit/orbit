/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package orbit

import (
	"net"
)

// ChannelBase is a basic channel implementation.
// Embed this to your channel struct and overwrite the AcceptStream
// method if required.
type ChannelBase struct {
	id string
	s  *Session
}

// NewChannelBase creates a new channel base with the given channel ID.
func NewChannelBase(id string) *ChannelBase {
	return &ChannelBase{
		id: id,
	}
}

// Init is called during channel registration.
func (c *ChannelBase) Init(s *Session) {
	c.s = s
}

// ID returns the channel ID.
func (c *ChannelBase) ID() string {
	return c.id
}

// Session returns the current active session of this channel.
func (c *ChannelBase) Session() *Session {
	return c.s
}

// OpenStream triggers a handshake to the peer and opens a new stream.
func (c *ChannelBase) OpenStream() (net.Conn, error) {
	return c.s.OpenStream(c.id)
}

// AcceptStream accepts new incoming streams from the peer.
// By default incoming streams are closed.
func (c *ChannelBase) AcceptStream(conn net.Conn) error {
	conn.Close()
	return ErrAcceptStreamsDisabled
}
