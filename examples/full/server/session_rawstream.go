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
	"net"
	"strings"
	"time"

	"github.com/desertbit/orbit/examples/full/api"
)

const orbitASCII = `
         ,MMM8&&&.
    _...MMMMM88&&&&..._
 .::'''MMMMM88&&&&&&'''::.
::     MMMMM88&&&&&&     ::
'::....MMMMM88&&&&&&....::'
    ''''MMMMM88&&&&''''
         'MMM8&&&'
`

// handleStreamOrbit is a showcase of the server side implementation of streaming on a
// raw net.Conn without using any helpers.
func handleStreamOrbit(stream net.Conn) error {
	// Ensure stream is closed.
	defer stream.Close()

	// Split the string, as we want to send one line at a time over the stream.
	orbitParts := strings.Split(orbitASCII, "\n")

	for i := 0; i < len(orbitParts); i++ {
		// For better output readability.
		time.Sleep(500 * time.Millisecond)

		// Set a write deadline.
		err := stream.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			return fmt.Errorf("error setting write deadline to stream '%s': %v", api.ChannelOrbit, err)
		}

		// Write one line directly onto the stream.
		n, err := stream.Write([]byte(orbitParts[i]))
		if err != nil {
			return fmt.Errorf("error writing to stream '%s': %v", api.ChannelOrbit, err)
		}
		if n != len(orbitParts[i]) {
			return fmt.Errorf("error writing to stream '%s': could only write %d bytes, expected to write %d bytes", api.ChannelOrbit, n, len(orbitParts[i]))
		}
	}

	return nil
}
