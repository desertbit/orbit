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

package main

import (
	"context"

	"github.com/desertbit/orbit/examples/simple/hello"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog/log"
)

var (
	_ hello.HelloClientHandler = &Client{}
	_ orbit.SessionHandler     = &Client{}
)

type Client struct {
	hello.HelloClientCaller
}

// Implements the orbit.SessionHandler interface.
func (c *Client) InitSession(s *orbit.Session) {
	c.HelloClientCaller = hello.RegisterHelloClient(s, c)
}

// Implements the api.HelloClientHandler interface.
func (c *Client) WhoAreYou(ctx context.Context, s *orbit.Session) (ret *hello.Info, err error) {
	log.Info().Msg("server asks who I am")
	ret = &hello.Info{
		Name:    "Max Mustermann",
		Age:     21,
		Locale:  "de_DE",
		Address: "Musterweg 14",
	}
	return
}
