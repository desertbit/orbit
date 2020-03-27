/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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

package service

import "context"

// A Context defines the service context which extends the context.Context interface.
type Context interface {
	context.Context

	// SetContext can be used to wrap the context.Context with additonal deadlines, ...
	SetContext(ctx context.Context)

	// Session returns the current active session.
	Session() Session

	// Header returns the raw header byte slice defined by the key. Returns nil if not present.
	Header(key string) []byte

	// Data returns the value defined by the key. Returns nil if not present.
	Data(key string) interface{}

	// SetData sets the value defined by the key.
	SetData(key string, v interface{})
}

type serviceContext struct {
	context.Context

	s      Session
	header map[string][]byte
	data   map[string]interface{}
}

func newContext(ctx context.Context, s Session, header map[string][]byte) Context {
	return &serviceContext{
		Context: ctx,
		s:       s,
		header:  header,
	}
}

func (c *serviceContext) SetContext(ctx context.Context) {
	c.Context = ctx
}

func (c *serviceContext) Session() Session {
	return c.s
}

func (c *serviceContext) Header(key string) []byte {
	return c.header[key]
}

func (c *serviceContext) Data(key string) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

func (c *serviceContext) SetData(key string, v interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = v
}
