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

package main

import (
	"log"

	"github.com/desertbit/orbit"
)

type CustomChannel struct {
	*orbit.ChannelBase
}

func NewCustomChannel(id string) *CustomChannel {
	return &CustomChannel{
		ChannelBase: orbit.NewChannelBase(id),
	}
}

func (c *CustomChannel) Open() (err error) {
	stream, err := c.OpenStream()
	if err != nil {
		return
	}

	// Hint: In a real scenario move this to a new goroutine.
	buf := make([]byte, 1000)
	n, err := stream.Read(buf)
	if err != nil {
		return
	}

	log.Println("custom channel: server message:", string(buf[:n]))

	return
}
