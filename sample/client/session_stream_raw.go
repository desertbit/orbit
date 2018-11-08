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
	"github.com/desertbit/orbit/sample/api"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// streamRawRoutine is a showcase of the client side implementation of streaming on a
// raw net.Conn without using any helpers.
func streamRawRoutine(stream net.Conn, wg *sync.WaitGroup) {
	defer func() {
		// For better output readability.
		fmt.Println("---------------------------")
		stream.Close()
		wg.Done()
	}()

	var (
		n   int
		err error
		buf = make([]byte, 256)
	)

	for {
		// Set a read timeout.
		err = stream.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Printf("error reading from stream '%s': could not set read deadline", api.ChannelIDRaw)
			return
		}

		// Read from the stream.
		n, err = stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("error reading from stream '%s': %v", api.ChannelIDRaw, err)
			}
			return
		}
		if n == 0 {
			log.Printf("error reading from stream '%s': no data read", api.ChannelIDRaw)
			return
		}

		// Write the data to the terminal.
		fmt.Println(string(buf[:n]))
	}
}
