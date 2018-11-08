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
	"log"
	"net"
	"sync"
	"time"
)

const (
	maxPayloadSize = 10 * 1024 // 10KB
)

// streamPacketRoutine is a showcase of the client side implementation of streaming data
// by using the packet pkg.
func streamPacketRoutine(stream net.Conn, wg *sync.WaitGroup) {
	defer func() {
		// For better output readability.
		fmt.Println("---------------------------")
		stream.Close()
		wg.Done()
	}()

	// For better output readability.
	time.Sleep(1 * time.Second)

	// Set a write deadline.
	err := stream.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Printf("error setting write deadline with packet: %v", err)
		return
	}

	// Write the first message to the server.
	err = packet.Write(stream, []byte("Packet 1: Hello, Server!"), maxPayloadSize)
	if err != nil {
		log.Printf("error writing to stream with packet: %v", err)
		return
	}

	// Read the witty response from the server.
	// Notice that we let packet set the read deadline on its own.
	data, err := packet.ReadTimeout(stream, nil, maxPayloadSize, 10*time.Second)
	if err != nil {
		log.Printf("error reading from stream with packet: %v", err)
		return
	}

	// Print the output
	fmt.Println(string(data))
}
