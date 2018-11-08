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

package signaler

import (
	"errors"
	"fmt"

	"github.com/desertbit/orbit/codec"
)

var (
	// ErrNoContextData is an error indicating that no context data is available.
	ErrNoContextData = errors.New("no context data available to decode")
)

// The Context type defines a signal context carrying the payload data
// that has been sent in the trigger request.
// It offers a convenience wrapper to the data, as it wraps its codec
// and offers easy access to the decoded data.
type Context struct {
	// Data is the raw byte representation of the encoded context data.
	Data []byte

	codec codec.Codec
}

// newContext creates a new context from the given data and its codec.
func newContext(data []byte, codec codec.Codec) *Context {
	return &Context{
		Data:  data,
		codec: codec,
	}
}

// Decode decodes the context data to a custom value using the internal codec.
// The value has to be passed as pointer.
// Returns ErrNoContextData if there is no context data available to decode.
func (c *Context) Decode(v interface{}) error {
	// Check if no data was passed.
	if len(c.Data) == 0 {
		return ErrNoContextData
	}

	// Decode the data into the value.
	err := c.codec.Decode(c.Data, v)
	if err != nil {
		return fmt.Errorf("decode: %v", err)
	}

	return nil
}
