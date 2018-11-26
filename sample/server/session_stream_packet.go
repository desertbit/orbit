/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
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
	"fmt"
	"github.com/desertbit/orbit/packet"
	"net"
	"time"
)

const (
	maxPayloadSize = 10 * 1024 // KB
)

// TODO: consider removal
// handleStreamPacket is a showcase of the server side implementation of streaming data
// by using the packet pkg.
func handleStreamPacket(stream net.Conn) error {
	// Set the read deadline manually
	err := stream.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return fmt.Errorf("error setting read deadline with packet: %v", err)
	}

	// Read data from stream.
	// We give nil for the buffer, therefore packet will allocate a buffer that fits the
	// payload of the message.
	data, err := packet.Read(stream, nil, maxPayloadSize)
	if err != nil {
		return fmt.Errorf("error reading from stream with packet: %v", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("error reading from stream with packet: no data read")
	}

	// Output the data to the console.
	fmt.Println(string(data))

	// For better output readability.
	time.Sleep(1 * time.Second)

	// Write a witty response.
	// This time, we let packet set the timeout.
	err = packet.WriteTimeout(stream, []byte("Packet 2: Hello, Client!"), maxPayloadSize, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error writing to stream with packet: %v", err)
	}

	return nil
}
