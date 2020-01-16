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
	"time"

	"github.com/desertbit/orbit/examples/simple/api"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog/log"
)

var _ api.ExampleConsumerHandler = &Client{}

type Client struct {
	api.ExampleConsumerCaller
}

func NewClient(co *orbit.Session) (caller api.ExampleConsumerCaller) {
	c := &Client{}
	c.ExampleConsumerCaller = api.RegisterExampleConsumer(co, c)

	caller = c
	return
}

// Implements the api.ExampleConsumerHandler interface.
func (c *Client) Test3(ctx context.Context, s *orbit.Session, args *api.ExampleTest3Args) (ret *api.ExampleTest3Ret, err error) {
	log.Info().Interface("args", args).Msg("Test3 Handler")
	ret = &api.ExampleTest3Ret{Lol: "not a dummy"}
	return
}

// Implements the api.ExampleConsumerHandler interface.
func (c *Client) Test4(ctx context.Context, s *orbit.Session) (ret *api.ExampleRect, err error) {
	ret = &api.ExampleRect{C: &api.ExampleChar{Lol: "not a dummy"}, X1: 1, X2: 1, Y1: 1, Y2: 1}
	return
}

// Implements the api.ExampleConsumerHandler interface.
func (c *Client) Hello3(s *orbit.Session, ret *api.PlateWriteChan) (err error) {
	for i := 0; i < 3; i++ {
		ret.C <- &api.Plate{
			Name:  "not a dummy",
			Rect:  &api.Rect{C: &api.Char{Lol: "not a dummy"}, X1: 1, X2: 1, Y1: 1, Y2: 1},
			Test:  map[int]*api.Rect{0: {C: &api.Char{Lol: "not a dummy"}, X1: 1, X2: 1, Y1: 1, Y2: 1}},
			Test2: []*api.Rect{{C: &api.Char{Lol: "not a dummy"}, X1: 1, X2: 1, Y1: 1, Y2: 1}},
			Test3: []float32{1, 2, 3},
			Test4: map[string]map[int][]*api.Rect{
				"Test": {
					0: []*api.Rect{{C: &api.Char{Lol: "not a dummy"}, X1: 1, X2: 1, Y1: 1, Y2: 1}},
				},
			},
			Ts:      time.Now(),
			Version: 5,
		}
	}
	ret.Close_()
	return
}

// Implements the api.ExampleConsumerHandler interface.
func (c *Client) Hello4(s *orbit.Session, args *api.ExampleCharReadChan, ret *api.PlateWriteChan) (err error) {
	go func() {
		for i := 0; i < 3; i++ {
			arg := <-args.C
			log.Info().Interface("arg", arg).Msg("Hello4 Handler")
		}
	}()
	for i := 0; i < 3; i++ {
		ret.C <- &api.Plate{
			Name:    "not a dummy",
			Ts:      time.Now(),
			Version: 5,
		}
	}
	ret.Close_()
	return
}
