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
	"context"
	"fmt"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/examples/simple/api"
	"github.com/desertbit/orbit/pkg/net/yamux"
	"github.com/desertbit/orbit/pkg/orbit"
	yamux2 "github.com/hashicorp/yamux"
	"github.com/rs/zerolog/log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("client run")
	}
}

func run() (err error) {
	cl := closer.New()
	defer cl.Close_()

	conn, err := yamux.NewTCPConn("127.0.0.1:6789", yamux2.DefaultConfig())
	if err != nil {
		return
	}

	co, err := orbit.NewClient(cl.CloserTwoWay(), conn, &orbit.Config{PrintPanicStackTraces: true})
	if err != nil {
		return
	}

	c := NewClient(co)

	// Make example calls.
	rect, err := c.Test(context.Background(), &api.Plate{Name: "PlateName", Rect: &api.Rect{X1: 2, X2: 3, Y1: 4, Y2: 5}})
	if err != nil {
		return
	}

	fmt.Printf("Test: %#v\n", rect)
	return
}
