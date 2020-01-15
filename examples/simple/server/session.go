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
	"net"

	"github.com/desertbit/orbit/examples/simple/api"
	"github.com/desertbit/orbit/pkg/orbit"
)

var _ api.ExampleProviderHandler = &Session{}

type Session struct {
	api.ExampleProviderCaller
}

// Implements the api.ExampleProviderHandler interface.
func (*Session) Test(ctx context.Context, s *orbit.Session, args *api.Plate) (ret *api.Rect, err error) {
	panic("implement me")
}

// Implements the api.ExampleProviderHandler interface.
func (*Session) Test2(ctx context.Context, s *orbit.Session, args *api.Rect) (err error) {
	panic("implement me")
}

// Implements the api.ExampleProviderHandler interface.
func (*Session) Hello(s *orbit.Session, stream net.Conn) (err error) {
	panic("implement me")
}

// Implements the api.ExampleProviderHandler interface.
func (*Session) Hello2(s *orbit.Session, args *api.CharReadChan) (err error) {
	panic("implement me")
}
