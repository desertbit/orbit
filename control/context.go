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

package control

import (
	"errors"
	"fmt"

	"github.com/desertbit/closer"
)

var (
	// ErrNoContextData defines the error if no context data is available.
	ErrNoContextData = errors.New("no context data available to decode")
)

// The Context type is a wrapper around the raw payload data of calls.
// It offer a convenience method to decode the encoded data into an
// interface.
type Context struct {
	// closer is used to signal to the handling func that the request
	// has been cancelled and that execution can be aborted.
	closer closer.Closer

	// ctrl is a reference to the Control the call has been issued on.
	ctrl *Control

	// Data is the raw byte representation of the encoded context data.
	Data []byte
}

// newContext creates a new Context from the given Control and the
// payload data.
func newContext(ctrl *Control, data []byte) *Context {
	return &Context{
		ctrl:   ctrl,
		closer: closer.New(),
		Data:   data,
	}
}

// Control returns the control of the context.
func (c *Context) Control() *Control {
	return c.ctrl
}

// Decode the context data to a custom value.
// The value has to be passed as pointer.
// Returns ErrNoContextData, if there is no context data available to decode.
// Returns ErrNoCodecAvailable, if there is no codec defined on the control.
func (c *Context) Decode(v interface{}) error {
	// Check if no data was passed.
	if len(c.Data) == 0 {
		return ErrNoContextData
	}

	// Decode the data.
	err := c.ctrl.config.Codec.Decode(c.Data, v)
	if err != nil {
		return fmt.Errorf("decode: %v", err)
	}

	return nil
}

// TODO:
func (c *Context) CancelChan() <-chan struct{} {
	return c.closer.CloseChan()
}
