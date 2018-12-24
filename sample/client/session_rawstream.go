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

package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/desertbit/orbit/sample/api"
)

// streamRawRoutine is a showcase of the client side implementation of streaming on a
// raw net.Conn without using any helpers.
func (s *Session) readStreamOrbit(stream net.Conn) {
	defer func() {
		// For better output readability.
		fmt.Println("---------------------------")
		_ = stream.Close()
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
			fmt.Printf("error reading from stream '%s': could not set read deadline", api.ChannelOrbit)
			return
		}

		// Read from the stream.
		n, err = stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("error reading from stream '%s': %v", api.ChannelOrbit, err)
			}
			return
		}
		if n == 0 {
			fmt.Printf("error reading from stream '%s': no data read", api.ChannelOrbit)
			return
		}

		// Write the data to the terminal.
		fmt.Println(string(buf[:n]))
	}
}
