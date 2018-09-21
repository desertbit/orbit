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
package control

import (
	"errors"
	"fmt"
)

var (
	// ErrNoContextData defines the error if no context data is available.
	ErrNoContextData = errors.New("no context data available to decode")
)

//####################//
//### Conext Type ####//
//####################//

// A Context defines a function context.
type Context struct {
	// Data is the raw byte representation of the encoded context data.
	Data []byte

	ctrl *Control
}

func newContext(ctrl *Control, data []byte) *Context {
	return &Context{
		ctrl: ctrl,
		Data: data,
	}
}

// Socket returns the socket of the context.
func (c *Context) Control() *Control {
	return c.ctrl
}

// Decode the context data to a custom value.
// The value has to be passed as pointer.
// Returns ErrNoContextData if there is no context data available to decode.
func (c *Context) Decode(v interface{}) error {
	// Check if no data was passed.
	if len(c.Data) == 0 {
		return ErrNoContextData
	}

	// Decode the data.
	err := c.ctrl.Codec.Decode(c.Data, v)
	if err != nil {
		return fmt.Errorf("decode: %v", err)
	}

	return nil
}
