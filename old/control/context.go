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
)

var (
	// ErrNoContextData defines the error if no context data is available.
	ErrNoContextData = errors.New("no context data available to decode")
)

// The Context type is a wrapper around the raw payload data of calls.
// It offers a convenience method to decode the encoded data into an
// interface and contains a closer that can be used to cancel the ongoing
// associated request.
type Context struct {
	// cancelChan is used to signal to the handling func that the request
	// has been canceled and that execution can be aborted.
	// Buffered channel of size 1.
	cancelChan chan struct{}

	// ctrl is a reference to the Control the call has been issued on.
	ctrl *Control

	// Data is the raw byte representation of the encoded context data.
	Data []byte
}

// newContext creates a new Context from the given Control and the
// payload data.
func newContext(ctrl *Control, data []byte, cancelable bool) *Context {
	c := &Context{
		ctrl: ctrl,
		Data: data,
	}
	if cancelable {
		c.cancelChan = make(chan struct{}, 1)
	}
	return c
}

// Control returns the control of the context.
func (c *Context) Control() *Control {
	return c.ctrl
}

// Decode the context data to a custom value.
// The value has to be passed as pointer.
// Returns ErrNoContextData, if there is no context data available to decode.
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

// CancelChan returns the cancel channel of the context.
// It can be used to detect whether a context has been canceled.
// Attention: Since the caller of a remote call can decide, whether his
// call is cancelable or not, the cancel channel of the context may be nil!
// Users of this method should never read directly from this channel,
// but rather use a select clause.
func (c *Context) CancelChan() <-chan struct{} {
	return c.cancelChan
}