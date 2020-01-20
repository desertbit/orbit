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
	}
	if n != 1 {
		return errors.New("failed to write flush byte to connection")
	}

	// Read the flush byte from the connection.
	n, err = conn.Read(b)
	if err != nil {
		return
	}
	if n != 1 {
		return errors.New("failed to read flush byte from connection")
	}
	if b[0] != flushByte {
		return errors.New("flush byte is invalid")
	}

	// At this point, we can be sure that all previous data has been flushed
	// and has reached the remote peer.

	// Reset the deadlines.
	return conn.SetDeadline(time.Time{})
}
