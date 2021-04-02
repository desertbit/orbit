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

package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

var (
	errFailed = errors.New("failed")
)

// T implenents the testify require and assert TestingT interface.
type T interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

type t struct {
	errors []error
	failed bool
}

func newT() *t {
	return &t{}
}

func (t *t) Errorf(format string, args ...interface{}) {
	t.errors = append(t.errors, fmt.Errorf(format, args...))
}

func (t *t) FailNow() {
	t.failed = true
	panic(errFailed)
}

// Tester is a concurrent test helper.
type Tester interface {
	Add(c uint)
	Run(func(context.Context, T))
	Wait(*testing.T)
}

type tester struct {
	ctx    context.Context
	cancel context.CancelFunc

	mx    *sync.Mutex
	count int
	cond  *sync.Cond

	tChan chan *t
}

func NewTester() Tester {
	ctx, cancel := context.WithCancel(context.Background())
	t := &tester{
		ctx:    ctx,
		cancel: cancel,
		mx:     &sync.Mutex{},
		tChan:  make(chan *t, 3),
	}
	t.cond = sync.NewCond(t.mx)
	return t
}

func (t *tester) Add(c uint) {
	t.mx.Lock()
	t.count += int(c)
	t.mx.Unlock()
}

func (t *tester) Run(f func(context.Context, T)) {
	gt := newT()

	defer func() {
		if err := recover(); err != nil && err != errFailed {
			gt.Errorf("catched panic: %v", err)
		}
		select {
		case t.tChan <- gt:
		case <-t.ctx.Done():
		}

		// Signal that this func has ended.
		t.mx.Lock()
		t.count--
		if t.count < 0 {
			panic("tester: count is negative, maybe you forgot to call Add()?")
		}
		t.cond.Broadcast()
		t.mx.Unlock()
	}()
	f(t.ctx, gt)
}

func (t *tester) Wait(tt *testing.T) {
	t.mx.Lock()
	defer t.mx.Unlock()

Loop:
	for {
		select {
		case gt := <-t.tChan:
			for _, err := range gt.errors {
				tt.Error(err)
			}
			if gt.failed {
				t.cancel()
				tt.FailNow()
				return
			}
		default:
			if t.count == 0 {
				break Loop
			}
			t.cond.Wait()
		}

	}
	t.cancel() // Always call cancel once done.
}
