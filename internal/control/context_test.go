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

package control_test

import (
	"errors"
	"testing"

	"github.com/desertbit/orbit/internal/control"
	"github.com/stretchr/testify/require"
)

func TestContext_Control(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(&control.Config{SendErrToCaller: true}, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callControl"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Check, if the Control is correctly set.
		if ctx.Control() != ctrl1 {
			err = errors.New("control was not set to the expected control")
		}
		return
	})
	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Check, if the Control is correctly set.
		if ctx.Control() != ctrl2 {
			err = errors.New("control was not set to the expected control")
		}
		return
	})

	_, err := ctrl1.Call(call, nil)
	require.NoError(t, err, "call 1: %v")

	_, err = ctrl2.Call(call, nil)
	require.NoError(t, err, "call 2: %v")
}

func TestContext_Decode(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(&control.Config{SendErrToCaller: true}, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callDecode"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		var test string
		err = ctx.Decode(&test)
		return
	})

	_, err := ctrl2.Call(call, nil)
	require.EqualError(t, err, control.ErrNoContextData.Error())

	_, err = ctrl2.Call(call, []byte{54, 51, 50, 1, 5, 2})
	require.Error(t, err, "expected decode to fail")
}
