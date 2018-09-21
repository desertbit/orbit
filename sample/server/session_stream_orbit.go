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
	"fmt"
	"github.com/desertbit/orbit/sample/api"
	"net"
	"strings"
	"time"
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

func handleOrbitStream(stream net.Conn) error {
	// Ensure stream is closed
	defer stream.Close()

	orbitParts := strings.Split(orbitASCII, "\n")

	for i := 0; i < len(orbitParts); i++ {
		time.Sleep(500 * time.Millisecond)

		n, err := stream.Write([]byte(orbitParts[i]))
		if err != nil {
			return fmt.Errorf("error writing to stream '%s': %v", api.ChannelIDOrbit, err)
		}
		if n != len(orbitParts[i]) {
			return fmt.Errorf("error writing to stream '%s': could only write %d bytes, expected to write %d bytes", api.ChannelIDOrbit, n, len(orbitParts[i]))
		}
	}

	return nil
}
