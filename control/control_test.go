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
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/desertbit/closer"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/packet"
	"github.com/pkg/errors"
)

var (
	defCtrl1, defCtrl2 *control.Control
	defPeer1, defPeer2 net.Conn
)

func TestMain(m *testing.M) {
	// Setup
	defPeer1, defPeer2 = net.Pipe()

	defCtrl1 = control.New(defPeer1, &control.Config{SendErrToCaller: true})
	defCtrl2 = control.New(defPeer2, &control.Config{SendErrToCaller: true})

	defCtrl1.Ready()
	defCtrl2.Ready()

	// Run
	code := m.Run()

	// Teardown
	err := defCtrl1.Close()
	if err != nil {
		log.Fatalf("defCtrl1 close: %v", err)
	}
	err = defCtrl2.Close()
	if err != nil {
		log.Fatalf("defCtrl2 close: %v", err)
	}

	// Exit
	os.Exit(code)
}

func TestControl_General(t *testing.T) {
	t.Parallel()

	assert(t, defCtrl1.LocalAddr() == defPeer1.LocalAddr(), "expected local addresses to be equal")
	assert(t, defCtrl1.RemoteAddr() == defPeer1.RemoteAddr(), "expected local addresses to be equal")
}

func TestControl_Call(t *testing.T) {
	t.Parallel()

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
	defCtrl1.AddFunc(call, f)
	defCtrl2.AddFunc(call, f)

	var ret string

	ctx, err := defCtrl1.Call(call, input)
	checkErr(t, "call 1", err)
	checkErr(t, "decode ret 1", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 1: expected '%v', got '%v'", output, ret)

	ctx, err = defCtrl2.Call(call, input)
	checkErr(t, "call 2", err)
	checkErr(t, "decode ret 2", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 2: expected '%v', got '%v'", output, ret)
}

func TestControl_CallOpts(t *testing.T) {
	t.Parallel()

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
		case <-time.After(10 * time.Millisecond):
			errChan <- errors.New("not cancelled")
		}
		return
	}

	const call = "callTimeout"
	const call2 = "callCancel"
	defCtrl1.AddFunc(call, fTimout)
	defCtrl2.AddFunc(call, f)
	defCtrl1.AddFunc(call2, fCancel)

	var ret string

	// Timeout not exceeded.
	ctx, err := defCtrl1.CallOpts(call, input, 100*time.Millisecond, nil)
	checkErr(t, "call 1", err)
	checkErr(t, "decode ret 1", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 1: expected '%v', got '%v'", output, ret)

	// Timeout exceeded.
	ctx, err = defCtrl2.CallOpts(call, input, 100*time.Millisecond, nil)
	assert(t, err == control.ErrCallTimeout, "call 2: expected '%v', got '%v'", control.ErrCallTimeout, err)

	// Cancel request.
	canceller := closer.New()
	// Cancel the request immediately.
	go func() {
		_ = canceller.Close()
	}()
	ctx, err = defCtrl2.CallOpts(call2, input, 100*time.Millisecond, canceller.CloseChan())
	checkErr(t, "call 2", err)
	assert(t, <-errChan == nil, "call cancel")
}

func TestControl_CallOneWay(t *testing.T) {
	t.Parallel()

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
	defCtrl1.AddFuncs(map[string]control.Func{call: f})
	defCtrl2.AddFuncs(map[string]control.Func{call: f})

	err := defCtrl1.CallOneWay(call, input)
	checkErr(t, "call 1", err)

	err = defCtrl2.CallOneWay(call, input)
	checkErr(t, "call 2", err)

	// Check the result of both one way calls.
	checkErr(t, "one way result", <-errChan)
	checkErr(t, "one way result", <-errChan)
}

func TestControl_CallAsync(t *testing.T) {
	t.Parallel()

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
	defCtrl1.AddFuncs(map[string]control.Func{call: f})
	defCtrl2.AddFuncs(map[string]control.Func{call: f})

	resChan := make(chan error)
	cb := func(ctx *control.Context, err error) {
		// We can not fail in another routine, therefore report whatever
		// happened back to the test thread by using the channel.
		defer func() {
			resChan <- err
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

	err := defCtrl1.CallAsync(call, input, cb)
	checkErr(t, "call 1", err)
	checkErr(t, "cb 1", <-resChan)

	err = defCtrl2.CallAsync(call, input, cb)
	checkErr(t, "call 2", err)
	checkErr(t, "cb 2", <-resChan)
}

func TestControl_CallAsyncOpts(t *testing.T) {
	t.Parallel()

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
		case <-time.After(10 * time.Millisecond):
			errChan <- errors.New("not cancelled")
		}
		return
	}

	const call = "callAsyncOpts"
	defCtrl1.AddFuncs(map[string]control.Func{call: f})
	defCtrl2.AddFunc(call, fCancel)

	resChan := make(chan error)
	cb := func(ctx *control.Context, err error) {
		// We can not fail in another routine, therefore report whatever
		// happened back to the test thread by using the channel.
		// We just expect a timeout error.
		resChan <- err
	}

	err := defCtrl2.CallAsyncOpts(call, nil, 10*time.Millisecond, cb, nil)
	checkErr(t, "call 1", err)
	err = <-resChan
	assert(t, err.Error() == control.ErrCallTimeout.Error(), "expected timeout err '%v', got '%v'", control.ErrCallTimeout, err)

	// Test a cancellation of the request.
	canceller := closer.New()
	_ = canceller.Close()
	err = defCtrl1.CallAsyncOpts(call, nil, 0, nil, canceller.CloseChan())
	checkErr(t, "call 2", err)
	assert(t, <-errChan == nil, "expected no error")
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
	assert(t, err.Error() == control.ErrClosed.Error(), "wrong error; expected '%v', got '%v'", control.ErrClosed, err)
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
	if cerr, ok := err.(*control.ErrorCode); ok {
		assert(t, cerr.Code == errCode, "wrong error code; expected '%d', got '%d'", errCode, cerr.Code)
		assert(t, cerr.Error() == msg, "wrong error; expected '%s', got '%s'", msg, cerr.Error())
	} else {
		t.Fatal("expected control error")
	}

	// Check the result of the hooks.
	checkErr(t, "callhook", <-callHookChan)
	err = <-errHookChan
	assert(t, err.Error() == errMsg, "wrong hook error; expected '%s', got '%v'", errMsg, err)

	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Test, whether a nil error is accepted.
		err = control.Err(nil, msg, errCode)
		return
	})

	_, err = ctrl1.Call(call, input)
	// Check if the error contains the correct code and message.
	if cerr, ok := err.(*control.ErrorCode); ok {
		assert(t, cerr.Code == errCode, "wrong error code; expected '%d', got '%d'", errCode, cerr.Code)
		assert(t, cerr.Error() == msg, "wrong error; expected '%s', got '%s'", msg, cerr.Error())
	} else {
		t.Fatal("expected control error")
	}
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
	assert(t, err == control.ErrWriteTimeout, "wrong error 1; expected '%v', got '%v'", control.ErrWriteTimeout, err)

	// Trigger a write timeout during CallAsync.
	err = ctrl2.CallAsync(call, "Hello this is a test case", nil)
	assert(t, err == control.ErrWriteTimeout, "wrong error 2; expected '%v', got '%v'", control.ErrWriteTimeout, err)
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
	assert(t, err == packet.ErrMaxPayloadSizeExceeded, "wrong error; expected '%v', got '%v'", packet.ErrMaxPayloadSizeExceeded, err)

	// Now, set the MaxMessageSize to a value, that allows our header, but
	// is too small for the payload.
	config2.MaxMessageSize = 511
	_, err = ctrl2.Call(call, make([]byte, 512))
	assert(t, err == packet.ErrMaxPayloadSizeExceeded, "wrong error; expected '%v', got '%v'", packet.ErrMaxPayloadSizeExceeded, err)

	// Now, set the MaxMessageSize to a value, that allows both header and payload.
	config2.MaxMessageSize = 512
	_, err = ctrl2.Call(call, make([]byte, 512))
	checkErr(t, "should be valid message size", err)

	// At last, set the MaxMessageSize on the receiving peer too low, which
	// must result in the connection being closed.
	config1.MaxMessageSize = 511
	_, err = ctrl2.Call(call, make([]byte, 512))
	netErrorString := "io: read/write on closed pipe"
	assert(t, err != nil && err.Error() == netErrorString, "expected error '%v', got '%v'", netErrorString, err)
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
	assert(t, err == control.ErrCallTimeout, "expected a timeout error due to panic in handler func")

	// Test a panic in the hooks.
	// Start with the call hook.
	ctrl2.SetCallHook(func(c *control.Control, funcID string, ctx *control.Context) {
		panic("test")
	})
	_, err = ctrl1.Call(call, nil)
	assert(t, err != nil, "expected a timeout error due to panic in call hook")
	// Reset the hook.
	ctrl2.SetCallHook(nil)

	// Now the error hook.
	ctrl2.SetErrorHook(func(c *control.Control, funcID string, err error) {
		panic("test")
	})
	_, err = ctrl1.Call(call, nil)
	assert(t, err != nil, "expected a timeout error due to panic in error hook")
}

func TestControl_HandleFuncNotExist(t *testing.T) {
	t.Parallel()

	conf := &control.Config{CallTimeout: 5 * time.Millisecond}
	ctrl1, ctrl2 := testControls(conf, conf)
	defer closeControls(ctrl1, ctrl2)

	_, err := ctrl1.Call("blabla", nil)
	assert(t, err != nil, "expected error since handler func is not defined")
}

func TestControl_CloseReadRoutine(t *testing.T) {
	t.Parallel()

	logger := bytes.Buffer{}
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, &control.Config{Logger: log.New(&logger, "", 0)})
	ctrl2 := control.New(peer2, nil)
	defer closeControls(ctrl1, ctrl2)

	// Close the control now, before the read routine is started.
	err := peer1.Close()
	checkErr(t, "closing peer1", err)

	// Start the read routine on the closed connection.
	ctrl1.Ready()

	// Give the routine some time to start
	time.Sleep(5 * time.Millisecond)

	// Check, if a log message has been written.
	assert(t, logger.Len() > 0, "expected a log message to be written")
}

// convenience
func assert(t *testing.T, condition bool, fmt string, args ...interface{}) {
	if !condition {
		t.Fatalf(fmt, args...)
	}
}

// convenience
func checkErr(t *testing.T, msg string, err error) {
	if err != nil {
		t.Fatal(msg, err)
	}
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
		_ = ctrl.Close()
	}
}
