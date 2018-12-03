/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package control_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/packet"
	"github.com/pkg/errors"
)

var (
	ctrl1, ctrl2 *control.Control
	peer1, peer2 net.Conn
)

func TestMain(m *testing.M) {
	// Setup
	peer1, peer2 = net.Pipe()

	ctrl1 = control.New(peer1, &control.Config{SendErrToCaller: true})
	ctrl2 = control.New(peer2, &control.Config{SendErrToCaller: true})

	ctrl1.Ready()
	ctrl2.Ready()

	// Run
	code := m.Run()

	// Teardown
	err := ctrl1.Close()
	if err != nil {
		log.Fatalf("ctrl1 close: %v", err)
	}
	err = ctrl2.Close()
	if err != nil {
		log.Fatalf("ctrl2 close: %v", err)
	}

	// Exit
	os.Exit(code)
}

func TestControl_General(t *testing.T) {
	t.Parallel()

	assert(t, ctrl1.LocalAddr() == peer1.LocalAddr(), "expected local addresses to be equal")
	assert(t, ctrl1.RemoteAddr() == peer1.RemoteAddr(), "expected local addresses to be equal")
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
	ctrl1.AddFunc(call, f)
	ctrl2.AddFunc(call, f)

	var ret string

	ctx, err := ctrl1.Call(call, input)
	checkErr(t, "call 1: %v", err)
	checkErr(t, "decode ret 1: %v", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 1: expected '%v', got '%v'", output, ret)

	ctx, err = ctrl2.Call(call, input)
	checkErr(t, "call 2: %v", err)
	checkErr(t, "decode ret 2: %v", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 2: expected '%v', got '%v'", output, ret)
}

func TestControl_CallTimeout(t *testing.T) {
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

	const call = "callTimeout"
	ctrl1.AddFunc(call, fTimout)
	ctrl2.AddFunc(call, f)

	var ret string

	// Timeout not exceeded.
	ctx, err := ctrl1.CallTimeout(call, input, 100*time.Millisecond)
	checkErr(t, "call 1: %v", err)
	checkErr(t, "decode ret 1: %v", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 1: expected '%v', got '%v'", output, ret)

	// Timeout exceeded.
	ctx, err = ctrl2.CallTimeout(call, input, 100*time.Millisecond)
	assert(t, err == control.ErrCallTimeout, "call 2: expected '%v', got '%v'", control.ErrCallTimeout, err)
}

func TestControl_CallOneWay(t *testing.T) {
	t.Parallel()

	input := "args"

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
		return
	}

	const call = "callOneWay"
	ctrl1.AddFuncs(map[string]control.Func{call: f})
	ctrl2.AddFuncs(map[string]control.Func{call: f})

	err := ctrl1.CallOneWay(call, input)
	checkErr(t, "call 1: %v", err)

	err = ctrl2.CallOneWay(call, input)
	checkErr(t, "call 2: %v", err)
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
	ctrl1.AddFuncs(map[string]control.Func{call: f})
	ctrl2.AddFuncs(map[string]control.Func{call: f})

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

	err := ctrl1.CallAsync(call, input, cb)
	checkErr(t, "call 1: %v", err)
	checkErr(t, "cb 1: %v", <-resChan)

	err = ctrl2.CallAsync(call, input, cb)
	checkErr(t, "call 2: %v", err)
	checkErr(t, "cb 2: %v", <-resChan)
}

func TestControl_CallAsyncTimeout(t *testing.T) {
	t.Parallel()

	f := func(ctx *control.Context) (data interface{}, err error) {
		time.Sleep(250 * time.Millisecond)
		return
	}

	const call = "callAsyncTimeout"
	ctrl1.AddFuncs(map[string]control.Func{call: f})

	resChan := make(chan error)
	cb := func(ctx *control.Context, err error) {
		// We can not fail in another routine, therefore report whatever
		// happened back to the test thread by using the channel.
		// We just expect a timeout error.
		resChan <- err
	}

	err := ctrl2.CallAsyncTimeout(call, nil, 10*time.Millisecond, cb)
	checkErr(t, "call 2: %v", err)
	err = <-resChan
	assert(t, err.Error() == control.ErrCallTimeout.Error(), "expected timeout err '%v', got '%v'", control.ErrCallTimeout, err)
}

func TestControl_ErrorClose(t *testing.T) {
	t.Parallel()

	// We need our own controls here, since we close a control.
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, nil)
	ctrl2 := control.New(peer2, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	ctrl1.Ready()
	ctrl2.Ready()

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

	// We need our own controls here to overwrite the config.
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, nil)
	ctrl2 := control.New(peer2, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

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

	ctrl1.Ready()
	ctrl2.Ready()

	errCode := 400
	errMsg := "error"
	msg := "msg"

	f := func(ctx *control.Context) (data interface{}, err error) {
		err = control.Err(errors.New(errMsg), msg, errCode)
		return
	}

	ctrl1.AddFunc(call, f)

	_, err := ctrl2.Call(call, input)
	// Check if the error contains the correct code and message.
	if cerr, ok := err.(*control.ErrorCode); ok {
		assert(t, cerr.Code == errCode, "wrong error code; expected '%d', got '%d'", errCode, cerr.Code)
		assert(t, cerr.Error() == msg, "wrong error; expected '%s', got '%s'", msg, cerr.Error())
	} else {
		t.Fatal("expected control error")
	}

	// Check the result of the hooks.
	checkErr(t, "callhook: %v", <-callHookChan)
	err = <-errHookChan
	assert(t, err.Error() == errMsg, "wrong hook error; expected '%s', got '%v'", errMsg, err)
}

func TestControl_ReadWriteTimeout(t *testing.T) {
	t.Parallel()

	// We need our own controls here to overwrite the config.
	peer1, peer2 := net.Pipe()
	// Set ridiculous timeouts to safely trigger them.
	ctrl1 := control.New(peer1, nil)
	ctrl2 := control.New(peer2, &control.Config{
		WriteTimeout: time.Nanosecond,
		//ReadTimeout: time.Nanosecond,
	})
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callReadWriteTimeout"

	ctrl1.Ready()
	ctrl2.Ready()

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
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, nil)
	config2 := &control.Config{}
	ctrl2 := control.New(peer2, config2)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callMaxMessageSize"

	ctrl1.Ready()
	ctrl2.Ready()

	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
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
}

func assert(t *testing.T, condition bool, fmt string, args ...interface{}) {
	if !condition {
		t.Fatalf(fmt, args...)
	}
}

func checkErr(t *testing.T, fmt string, err error) {
	if err != nil {
		t.Fatalf(fmt, err)
	}
}
