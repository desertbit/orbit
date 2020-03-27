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

package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

type Client interface {
	closer.Closer

	Call(ctx context.Context, id string, arg, ret interface{}) error
	AsyncCall(ctx context.Context, id string, arg, ret interface{}) error
	Stream(ctx context.Context, id string) (transport.Stream, error)
}

type client struct {
	closer.Closer

	opts  *Options
	log   *zerolog.Logger
	hooks Hooks

	sessionMx          sync.Mutex
	session            *session
	connectSessionChan chan chan *session
}

func New(opts *Options) (Client, error) {
	opts.setDefaults()
	err := opts.validate()
	if err != nil {
		return nil, err
	}

	c := &client{
		Closer:             opts.Closer,
		opts:               opts,
		log:                opts.Log,
		hooks:              opts.Hooks,
		connectSessionChan: make(chan chan *session),
	}
	c.OnClose(c.hookClose)
	c.startSessionRoutine()
	return c, nil
}

func (c *client) Call(ctx context.Context, id string, arg, ret interface{}) error {
	if c.IsClosing() {
		return ErrClosed
	}

	// Get the connected session or trigger a connect attempt.
	s, err := c.connectedSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connected session: %w", err)
	}

	return s.Call(ctx, id, arg, ret)
}

func (c *client) AsyncCall(ctx context.Context, id string, arg, ret interface{}) error {
	if c.IsClosing() {
		return ErrClosed
	}

	// Get the connected session or trigger a connect attempt.
	s, err := c.connectedSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connected session: %w", err)
	}

	return s.AsyncCall(ctx, id, arg, ret)
}

func (c *client) Stream(ctx context.Context, id string) (transport.Stream, error) {
	if c.IsClosing() {
		return nil, ErrClosed
	}

	// Get the connected session or trigger a connect attempt.
	s, err := c.connectedSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected session: %w", err)
	}

	return s.OpenStream(ctx, id)
}
