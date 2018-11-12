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

/*
Package flusher provides convenience methods to flush a net.Conn.
*/
package flusher

import (
	"errors"
	"net"
	"time"
)

const (
	flushByte = 123
)

// Flush a connection by waiting until all data was written to the peer.
func Flush(conn net.Conn, timeout time.Duration) (err error) {
	// Set the read and write deadlines.
	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	// Write the flush byte to the connection.
	b := []byte{flushByte}
	n, err := conn.Write(b)
	if err != nil {
		return err
	} else if n != 1 {
		return errors.New("failed to write flush byte to connection")
	}

	// Read the flush byte from the connection.
	n, err = conn.Read(b)
	if err != nil {
		return
	} else if n != 1 {
		return errors.New("failed to read flush byte from connection")
	} else if b[0] != flushByte {
		return errors.New("flush byte is invalid")
	}

	// Reset the deadlines.
	return conn.SetDeadline(time.Time{})
}
