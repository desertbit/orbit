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
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/packet"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestControl_General(t *testing.T) {
	t.Parallel()

	// Create controls.
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, nil)
	ctrl2 := control.New(peer2, nil)

	ctrl1.Ready()
	ctrl2.Ready()

	defer closeControls(ctrl1, ctrl2)

	require.Equal(t, ctrl1.LocalAddr(), peer1.LocalAddr(), "expected local addresses to be equal")
	require.Equal(t, ctrl1.RemoteAddr(), peer1.RemoteAddr(), "expected remote addresses to be equal")
}

func TestControl_Call(t *testing.T) {
	t.Parallel()

	// Create controls.
	ctrl1, ctrl2 := testControls(
		&control.Config{SendErrToCaller: true},
		&control.Config{SendErrToCaller: true},
	)
	defer closeControls(ctrl1, ctrl2)

	input := "args"
	output := "ret"

	f := func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args: expected '%v', got '%v'", input, args)
			return
		}

		return output, nil
	}

	const call = "call"
	ctrl1.AddFunc(call, f)
	ctrl2.AddFunc(call, f)

	var ret string

	ctx, err := ctrl1.Call(call, input)
	require.NoError(t, err)
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, output, ret)

	ctx, err = ctrl2.Call(call, input)
	require.NoError(t, err)
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, output, ret)
}

func TestControl_CallOpts(t *testing.T) {
	t.Parallel()

	// Create controls.
	ctrl1, ctrl2 := testControls(
		&control.Config{SendErrToCaller: true},
		&control.Config{SendErrToCaller: true},
	)
	defer closeControls(ctrl1, ctrl2)

	input := "args"
	output := "ret"

	f := func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args: expected '%v', got '%v'", input, args)
			return
		}

		data = output
		return
	}

	// Exceed timeout.
	fTimout := func(ctx *control.Context) (data interface{}, err error) {
		time.Sleep(1 * time.Second)
		return
	}

	// Cancel.
	errChan := make(chan error)
	fCancel := func(ctx *control.Context) (data interface{}, err error) {
		select {
		case <-ctx.CancelChan():
			errChan <- nil
		case <-time.After(50 * time.Millisecond):
			errChan <- errors.New("not canceled")
		}
		return
	}

	const call = "callTimeout"
	const call2 = "callCancel"
	ctrl1.AddFunc(call, fTimout)
	ctrl2.AddFunc(call, f)
	ctrl1.AddFunc(call2, fCancel)

	var ret string

	// Timeout not exceeded.
	ctx, err := ctrl1.CallOpts(call, input, 100*time.Millisecond, nil)
	require.NoError(t, err)
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, output, ret)

	// Timeout exceeded.
	ctx, err = ctrl2.CallOpts(call, input, 100*time.Millisecond, nil)
	require.Exactly(t, control.ErrCallTimeout, err)

	// Cancel request.
	canceller := closer.New()
	// Cancel the request immediately.
	go func() {
		_ = canceller.Close()
	}()
	ctx, err = ctrl2.CallOpts(call2, input, 100*time.Millisecond, canceller.ClosingChan())
	require.Equal(t, control.ErrCallCanceled, err)
}

func TestControl_CallOneWay(t *testing.T) {
	t.Parallel()

	// Create controls.
	ctrl1, ctrl2 := testControls(
		&control.Config{SendErrToCaller: true},
		&control.Config{SendErrToCaller: true},
	)
	defer closeControls(ctrl1, ctrl2)

	input := "args"

	errChan := make(chan error)
	f := func(ctx *control.Context) (data interface{}, err error) {
		defer func() {
			// Report the result back.
			errChan <- err
		}()

		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args: expected '%v', got '%v'", input, args)
			return
		}
		return
	}

	const call = "callOneWay"
	ctrl1.AddFuncs(map[string]control.Func{call: f})
	ctrl2.AddFuncs(map[string]control.Func{call: f})

	err := ctrl1.CallOneWay(call, input)
	require.NoError(t, err)

	err = ctrl2.CallOneWay(call, input)
	require.NoError(t, err)

	// Check the result of both one way calls.
	require.NoError(t, <-errChan)
	require.NoError(t, <-errChan)
}

func TestControl_CallAsync(t *testing.T) {
	t.Parallel()

	// Create controls.
	ctrl1, ctrl2 := testControls(
		&control.Config{SendErrToCaller: true},
		&control.Config{SendErrToCaller: true},
	)
	defer closeControls(ctrl1, ctrl2)

	input := "args"
	output := "ret"

	f := func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args: expected '%v', got '%v'", input, args)
			return
		}

		data = output
		return
	}

	const call = "callAsync"
	ctrl1.AddFuncs(map[string]control.Func{call: f})
	ctrl2.AddFuncs(map[string]control.Func{call: f})

	errChan := make(chan error)
	cb := func(ctx *control.Context, err error) {
		// We can not fail in another routine, therefore report whatever
		// happened back to the test thread by using the channel.
		defer func() {
			errChan <- err
		}()
		if err != nil {
			return
		}

		var ret string
		err = ctx.Decode(&ret)
		if err != nil {
			return
		}

		if ret != output {
			err = errors.Errorf("decoded ret 1: expected '%v', got '%v'", output, ret)
			return
		}

		return
	}

	err := ctrl1.CallAsync(call, input, cb)
	require.NoError(t, err)
	require.NoError(t, <-errChan)

	err = ctrl2.CallAsync(call, input, cb)
	require.NoError(t, err)
	require.NoError(t, <-errChan)
}

func TestControl_CallAsyncOpts(t *testing.T) {
	t.Parallel()

	// Create controls.
	ctrl1, ctrl2 := testControls(
		&control.Config{SendErrToCaller: true},
		&control.Config{SendErrToCaller: true},
	)
	defer closeControls(ctrl1, ctrl2)

	f := func(ctx *control.Context) (data interface{}, err error) {
		time.Sleep(250 * time.Millisecond)
		return
	}

	// Cancel.
	errChan := make(chan error)
	fCancel := func(ctx *control.Context) (data interface{}, err error) {
		select {
		case <-ctx.CancelChan():
			errChan <- nil
		case <-time.After(50 * time.Millisecond):
			errChan <- errors.New("not canceled")
		}
		return
	}

	const call = "callAsyncOpts"
	ctrl1.AddFuncs(map[string]control.Func{call: f})
	ctrl2.AddFunc(call, fCancel)

	cb := func(ctx *control.Context, err error) {
		errChan <- err
	}

	err := ctrl2.CallAsyncOpts(call, nil, 10*time.Millisecond, cb, nil)
	require.NoError(t, err)
	require.EqualError(t, <-errChan, control.ErrCallTimeout.Error())

	// Test a cancellation of the request.
	canceller := closer.New()
	err = ctrl1.CallAsyncOpts(call, nil, 10*time.Millisecond, cb, canceller.ClosingChan())
	require.NoError(t, canceller.Close())
	require.NoError(t, err)
	require.Equal(t, control.ErrCallCanceled, <-errChan)
}

func TestControl_ErrorClose(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(nil, nil)
	defer closeControls(ctrl1, ctrl2)

	const call = "callErrorClose"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		time.Sleep(100 * time.Millisecond)
		return
	})

	time.AfterFunc(10*time.Millisecond, func() {
		_ = ctrl1.Close()
	})
	_, err := ctrl2.Call(call, nil)
	require.EqualError(t, err, control.ErrClosed.Error())
}

func TestControl_Error_And_Hooks(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(nil, nil)
	defer closeControls(ctrl1, ctrl2)

	const call = "callError"
	callHookChan := make(chan error, 1)
	input := "args"

	ctrl1.SetCallHook(func(c *control.Control, funcID string, ctx *control.Context) {
		var err error
		// Report the result back over the channel.
		defer func() {
			callHookChan <- err
		}()

		if ctrl1 != c {
			err = errors.New("controls were different")
			return
		}
		if funcID != call {
			err = fmt.Errorf("funcID wrong; expected '%s', got '%s'", call, funcID)
			return
		}
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = errors.WithMessage(err, "callhook")
			return
		}

		if args != input {
			err = fmt.Errorf("context contained incorrect data; expected '%s', got '%s'", input, args)
			return
		}
	})

	errHookChan := make(chan error, 1)

	ctrl1.SetErrorHook(func(c *control.Control, funcID string, err error) {
		// Report the result back over the channel.
		defer func() {
			errHookChan <- err
		}()
		if ctrl1 != c {
			err = errors.New("controls were different")
			return
		}
		if funcID != call {
			err = fmt.Errorf("funcID wrong; expected '%s', got '%s'", call, funcID)
			return
		}
	})

	errCode := 400
	errMsg := "error"
	msg := "msg"

	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Test a default control error.
		err = control.Err(errors.New(errMsg), msg, errCode)
		return
	})

	_, err := ctrl2.Call(call, input)
	// Check if the error contains the correct code and message.
	require.IsType(t, (*control.ErrorCode)(nil), err)
	cerr, _ := err.(*control.ErrorCode)
	require.Equal(t, errCode, cerr.Code)
	require.EqualError(t, cerr, msg)

	// Check the result of the hooks.
	require.NoError(t, <-callHookChan)
	require.EqualError(t, <-errHookChan, errMsg)

	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Test, whether a nil error is accepted.
		err = control.Err(nil, msg, errCode)
		return
	})

	_, err = ctrl1.Call(call, input)
	// Check if the error contains the correct code and message.
	require.IsType(t, (*control.ErrorCode)(nil), err)
	cerr, _ = err.(*control.ErrorCode)
	require.Equal(t, errCode, cerr.Code)
	require.EqualError(t, cerr, msg)
}

func TestControl_WriteTimeout(t *testing.T) {
	t.Parallel()

	// Set ridiculous timeouts to safely trigger them.
	ctrl1, ctrl2 := testControls(nil, &control.Config{WriteTimeout: time.Nanosecond})
	defer closeControls(ctrl1, ctrl2)

	const call = "callReadWriteTimeout"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		time.Sleep(time.Microsecond)
		return
	})

	// Trigger a write timeout during Call.
	_, err := ctrl2.Call(call, "Hello this is a test case")
	require.Equal(t, control.ErrWriteTimeout, err)

	// Trigger a write timeout during CallAsync.
	err = ctrl2.CallAsync(call, "Hello this is a test case", nil)
	require.Equal(t, control.ErrWriteTimeout, err)
}

func TestControl_MaxMessageSize(t *testing.T) {
	t.Parallel()

	// We need our own controls here to overwrite the config.
	config1 := &control.Config{}
	config2 := &control.Config{}
	ctrl1, ctrl2 := testControls(config1, config2)
	defer closeControls(ctrl1, ctrl2)

	const call = "callMaxMessageSize"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		return
	})
	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		return
	})

	// We set the message size at this point so small, that the header should
	// fail to be sent.
	config2.MaxMessageSize = 1
	_, err := ctrl2.Call(call, nil)
	require.Equal(t, packet.ErrMaxPayloadSizeExceeded, err)

	// Now, set the MaxMessageSize to a value, that allows our header, but
	// is too small for the payload.
	config2.MaxMessageSize = 511
	_, err = ctrl2.Call(call, make([]byte, 512))
	require.Equal(t, packet.ErrMaxPayloadSizeExceeded, err)

	// Now, set the MaxMessageSize to a value, that allows both header and payload.
	config2.MaxMessageSize = 512
	_, err = ctrl2.Call(call, make([]byte, 512))
	require.NoError(t, err, "should be valid message size")

	// At last, set the MaxMessageSize on the receiving peer too low, which
	// must result in the connection being closed.
	config1.MaxMessageSize = 511
	_, err = ctrl2.Call(call, make([]byte, 512))
	netErrorString := "io: read/write on closed pipe"
	require.EqualError(t, err, netErrorString)
}

func TestControl_Panic(t *testing.T) {
	t.Parallel()

	conf := &control.Config{CallTimeout: 5 * time.Millisecond}
	ctrl1, ctrl2 := testControls(conf, conf)
	defer closeControls(ctrl1, ctrl2)

	const call = "callPanic"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Cause a panic (err is nil).
		panic("test")
	})
	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Return an error to trigger the error hook.
		return nil, errors.New("test")
	})

	// Test a panic in a handler func.
	_, err := ctrl2.Call(call, nil)
	require.Equal(t, control.ErrCallTimeout, err)

	// Test a panic in the hooks.
	// Start with the call hook.
	ctrl2.SetCallHook(func(c *control.Control, funcID string, ctx *control.Context) {
		panic("test")
	})
	_, err = ctrl1.Call(call, nil)
	require.Equal(
		t,
		control.ErrCallTimeout, err,
		"expected call timeout due to panic in call hook",
	)
	// Reset the hook.
	ctrl2.SetCallHook(nil)

	// Now the error hook.
	ctrl2.SetErrorHook(func(c *control.Control, funcID string, err error) {
		panic("test")
	})
	_, err = ctrl1.Call(call, nil)
	require.Equal(
		t,
		control.ErrCallTimeout, err,
		"expected call timeout due to panic in error hook",
	)
}

func TestControl_HandleFuncNotExist(t *testing.T) {
	t.Parallel()

	conf := &control.Config{CallTimeout: 5 * time.Millisecond}
	ctrl1, ctrl2 := testControls(conf, conf)
	defer closeControls(ctrl1, ctrl2)

	_, err := ctrl1.Call("blabla", nil)
	require.Error(t, err, "expected error since handler func is not defined")
}

// convenience
func testControls(conf1, conf2 *control.Config) (ctrl1, ctrl2 *control.Control) {
	peer1, peer2 := net.Pipe()
	ctrl1 = control.New(peer1, conf1)
	ctrl2 = control.New(peer2, conf2)

	ctrl1.Ready()
	ctrl2.Ready()
	return
}

// convenience
func closeControls(ctrls ...*control.Control) {
	for _, ctrl := range ctrls {
		ctrl.Close_()
	}
}
