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
	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/examples/simple/api"
	"github.com/desertbit/orbit/pkg/orbit"
)

var _ api.ExampleConsumerHandler = &Client{}

type Client struct {
	api.ExampleConsumerCaller
}

func NewClient(co *orbit.Session) (caller api.ExampleConsumerCaller, err error) {
	c := &Client{}
	c.ExampleConsumerCaller, err = api.RegisterExampleConsumer(co, c)
	if err != nil {
		return
	}

	caller = c
	return
}

func (c *Client) Test3(args *api.Test3Args) (ret *api.Test3Ret, err error) {
	panic("implement me")
}

func (c *Client) Test4() (ret *api.Rect, err error) {
	panic("implement me")
}

func (c *Client) Hello3() (ret <-chan *api.Plate, err error) {
	panic("implement me")
}

func (c *Client) Hello4(args <-chan *api.Char) (ret <-chan *api.Plate, err error) {
	panic("implement me")
}

func main() {
	cl := closer.New()

	var conn orbit.Conn
	co, err := orbit.NewClient(cl.CloserTwoWay(), conn, &orbit.Config{PrintPanicStackTraces: true})
	if err != nil {
		return
	}

	c, err := NewClient(co)
	if err != nil {
		return
	}
}
