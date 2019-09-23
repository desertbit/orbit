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

package signaler_test

import (
	"bytes"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/packet"
	"github.com/desertbit/orbit/signaler"
	"github.com/stretchr/testify/require"
)

func TestSignaler_OnSignal(t *testing.T) {
	t.Parallel()

	const signal = "onSignal"
	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)
	data := "test data"

	sig1.AddSignal(signal)
	ln2 := sig2.OnSignal(signal)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Trigger the signal with some test data.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Read the signal off of the listener.
	ctx := <-ln2.C
	var ret string
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, data, ret)
}

func TestSignaler_MaxMessageSize(t *testing.T) {
	t.Parallel()

	const signal = "maxMessageSize"
	sig1, sig2 := testSignaler(&control.Config{MaxMessageSize: 100}, nil)
	defer closeSignalers(sig1, sig2)
	data := make([]byte, 15000)

	sig1.AddSignal(signal)
	_ = sig2.OnSignal(signal)

	// IMPORTANT! Wait a short time to safely activate signal.
	time.Sleep(time.Millisecond * 2)

	// Now trigger the signal of sig1, which should exceed the configured MaxMessageSize.
	err := sig1.TriggerSignal(signal, data)
	require.Equal(t, packet.ErrMaxPayloadSizeExceeded, err)
}

func TestSignaler_OnSignalOpts(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "onSignalOpts"
	sig1.AddSignal(signal)

	// First, test invalid channel sizes.
	require.Panics(t, func() {
		_ = sig2.OnSignalOpts(signal, -1)
	})
	require.Panics(t, func() {
		_ = sig2.OnSignalOpts(signal, 0)
	})

	// Now use a valid chan size.
	const chanSize = 3
	ln2 := sig2.OnSignalOpts(signal, chanSize)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	data := "test data"

	// Trigger the signals with some test data, but more than fit into the chan.
	for i := 0; i < chanSize*2; i++ {
		require.NoErrorf(t, sig1.TriggerSignal(signal, data), "trigger %d", i)
	}

	// Read the signals off of the listener.
	for i := 0; i < chanSize*2; i++ {
		select {
		case ctx := <-ln2.C:
			var ret string
			require.NoErrorf(t, ctx.Decode(&ret), "decode %d", i)
			require.Equalf(t, data, ret, "equal %d", i)
		case <-time.After(time.Millisecond * 10):
			t.Fatalf("timeout when waiting for events in OnSignalOpts, %d", i)
		}
	}
}

func TestSignaler_OnSignalFunc(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "onSignalFunc"
	const signalPanic = "onSignalFuncPanic"
	data := "test data"
	sig1.AddSignals([]string{signal, signalPanic})
	sig2.AddSignal(signalPanic)

	errChan := make(chan error)
	retChan := make(chan string)
	_ = sig2.OnSignalFunc(signal, func(ctx *signaler.Context) {
		var ret string
		err := ctx.Decode(&ret)
		if err != nil {
			errChan <- err
		} else {
			retChan <- ret
		}
	})
	_ = sig2.OnSignalFunc(signalPanic, func(ctx *signaler.Context) {
		panic("test")
	})
	_ = sig1.OnSignalFunc(signalPanic, nil)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Trigger the signal with some test data.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Wait for the result.
	select {
	case err := <-errChan:
		require.NoError(t, err)
	case ret := <-retChan:
		require.Equal(t, data, ret)
	}

	// Test a panic in the handler func.
	require.NoError(t, sig1.TriggerSignal(signalPanic, data))
	// Sleep shortly so that the signal has time to arrive.
	time.Sleep(time.Millisecond * 10)

	// Test a nil handler func.
	require.NoError(t, sig2.TriggerSignal(signalPanic, data))
}

func TestSignaler_OnceSignal(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "onceSignal"
	sig1.AddSignals([]string{signal})

	ln2 := sig2.OnceSignal(signal)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Trigger the signal with some test data.
	data := "test data"
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Read the signal off of the listener.
	ctx := <-ln2.C
	var ret string
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, data, ret)

	// Now trigger it again, but this time nothing should arrive.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Check that no signal has been triggered.
	select {
	case _ = <-ln2.C:
		t.Fatal("signal should only fire once")
	case <-time.After(time.Millisecond * 10):
	}
}

func TestSignaler_OnceSignalOpts(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "onceSignalOpts"
	sig1.AddSignal(signal)

	// First, test invalid channel sizes.
	// First, test invalid channel sizes.
	require.Panics(t, func() {
		_ = sig2.OnceSignalOpts(signal, -1)
	})
	require.Panics(t, func() {
		_ = sig2.OnceSignalOpts(signal, 0)
	})

	// Now use a valid chan size.
	const chanSize = 1
	ln2 := sig2.OnceSignalOpts(signal, chanSize)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	data := "test data"

	// Trigger the signals with some test data.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Read the signal off of the listener.
	select {
	case ctx := <-ln2.C:
		var ret string
		require.NoError(t, ctx.Decode(&ret))
		require.Equal(t, data, ret)
	case <-time.After(time.Millisecond * 10):
		t.Fatal("timeout when waiting for events in OnceSignalOpts")
	}
}

func TestSignaler_OnceSignalFunc(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "onceSignalFunc"
	data := "test data"
	sig1.AddSignal(signal)

	errChan := make(chan error)
	retChan := make(chan string)
	_ = sig2.OnceSignalFunc(signal, func(ctx *signaler.Context) {
		var ret string
		err := ctx.Decode(&ret)
		if err != nil {
			errChan <- err
		} else {
			retChan <- ret
		}
	})

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Trigger the signal with some test data.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Wait for the result.
	select {
	case err := <-errChan:
		require.NoError(t, err)
	case ret := <-retChan:
		require.Equal(t, data, ret)
	}

	// Trigger the signal again, but this time no signal should be triggered.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Wait for the result.
	select {
	case err := <-errChan:
		require.NoError(t, err)
	case res := <-retChan:
		require.Failf(t, "did not expect a result", "result: %v", res)
	case <-time.After(time.Millisecond * 10):
	}
}

func TestSignaler_SignalFilter(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer closeSignalers(sig1, sig2)

	const signal = "signalFilter"
	type filterData struct {
		ID int
	}
	type signalData struct {
		ID int
	}

	sig1.AddSignalFilter(signal, func(ctx *signaler.Context) (f signaler.Filter, err error) {
		var filData filterData
		err = ctx.Decode(&filData)
		if err != nil {
			return
		}

		f = func(data interface{}) (conforms bool, err error) {
			sigData, ok := data.(*signalData)
			if !ok {
				err = errors.New("cast to signalData failed")
				return
			}

			conforms = sigData.ID == filData.ID
			return
		}
		return
	})

	ln2 := sig2.OnSignal(signal)
	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Set a filter on the signal.
	require.NoError(t, sig2.SetSignalFilter(signal, &filterData{ID: 1}))

	// Trigger the signal with some test data that should not pass the filter.
	require.NoError(t, sig1.TriggerSignal(signal, &signalData{ID: 0}))

	// Check that no signal has been triggered.
	select {
	case _ = <-ln2.C:
		t.Fatal("signal should have been filtered out")
	case <-time.After(time.Millisecond * 10):
	}

	// Trigger the signal with some test data that should pass the filter.
	require.NoError(t, sig1.TriggerSignal(signal, &signalData{ID: 1}))

	// Read the signal off of the listener.
	ctx := <-ln2.C
	var ret signalData
	require.NoError(t, ctx.Decode(&ret))
	require.Equal(t, 1, ret.ID)
}

func TestSignaler_SignalErrors(t *testing.T) {
	t.Parallel()

	buffer := bytes.Buffer{}
	sig1, sig2 := testSignaler(
		&control.Config{Logger: log.New(&buffer, "", 0)},
		nil,
	)
	defer closeSignalers(sig1, sig2)

	// Register a signal multiple times! No error should happen,
	// but we expect a log to be printed.
	sig1.AddSignal("test3")
	sig1.AddSignal("test3")
	require.True(t, buffer.Len() > 0)

	// Empty the log buffer.
	buffer.Reset()
	sig1.AddSignals([]string{"test4", "test4"})
	require.True(t, buffer.Len() > 0)

	// Trigger a signal with nil data and try to decode it in the handler
	errChan := make(chan error)
	sig2.OnSignalFunc("test3", func(ctx *signaler.Context) {
		var args string
		t.Log(string(ctx.Data))
		errChan <- ctx.Decode(&args)
	})
	// Wait some time so that the signal state can be set.
	time.Sleep(2 * time.Millisecond)
	require.NoError(t, sig1.TriggerSignal("test3", nil))
	err := <-errChan
	require.Equal(t, signaler.ErrNoContextData, err)

	// Try to register a listener for a non-existent signal.
	lntmp := sig2.OnSignal("test")
	// Wait some time so that the signal state can be set.
	time.Sleep(2 * time.Millisecond)
	lntmp.Off()

	// Try to register a listener for a non-existent signal.
	// And immediately switch the listener off.
	lntmp2 := sig1.OnSignal("test2")
	lntmp2.Off()

	// Try to set a filter for a non-existent signal.
	err = sig2.SetSignalFilter("test", "test")
	require.Equal(t, signaler.ErrSignalNotFound, err)

	// Trigger a non-existent signal.
	err = sig1.TriggerSignal("test", nil)
	require.Equal(t, signaler.ErrSignalNotFound, err)

	// Try to set a filter on a signal, that does not allow filters.
	sig1.AddSignal("test")
	err = sig2.SetSignalFilter("test", "test")
	require.Equal(t, signaler.ErrFilterFuncUndefined, err)

	const signal = "triggerSignalError"
	// Add the signal and register listener for it.
	sig1.AddSignal(signal)
	ln := sig2.OnSignal(signal)
	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)
	// Deactivate the listener now, making the signal inactive.
	ln.Off()

	// Trigger the inactive signal.
	require.NoError(t, sig1.TriggerSignal(signal, nil))
}

// convenience
func testSignaler(conf1, conf2 *control.Config) (sig1, sig2 *signaler.Signaler) {
	peer1, peer2 := net.Pipe()
	sig1 = signaler.New(peer1, conf1)
	sig2 = signaler.New(peer2, conf2)

	sig1.Ready()
	sig2.Ready()
	return
}

// convenience
func closeSignalers(sigs ...*signaler.Signaler) {
	for _, sig := range sigs {
		_ = sig.Close()
	}
}
