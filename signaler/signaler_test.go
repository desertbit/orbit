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

package signaler_test

import (
	"net"
	"testing"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/packet"
	"github.com/desertbit/orbit/signaler"
)

func TestSignaler_OnSignal(t *testing.T) {
	t.Parallel()

	conf2 := &control.Config{SendErrToCaller: true}
	sig1, sig2 := testSignaler(
		&control.Config{SendErrToCaller: true},
		conf2,
	)
	defer func() {
		_ = sig1.Close()
		_ = sig2.Close()
	}()

	const signal = "onSignal"
	sig1.AddSignal(signal)
	sig2.AddSignal(signal)
	_ = sig1.OnSignal(signal)
	ln2 := sig2.OnSignal(signal)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Trigger the signal with some test data.
	data := "test data"
	err := sig1.TriggerSignal(signal, data)
	checkErr(t, "trigger signal 1", err)

	// Read the signal off of the listener.
	ctx := <-ln2.C
	var ret string
	err = ctx.Decode(&ret)
	checkErr(t, "decoding return data 1", err)
	assert(t, ret == data, "returned data wrong; expected '%v', got '%v'", data, ret)

	// Now trigger the signal of sig2, which should exceed the configured
	// MaxMessageSize.
	// Trigger the signal with some test data.
	conf2.MaxMessageSize = 1
	err = sig2.TriggerSignal(signal, data)
	assert(t, err == packet.ErrMaxPayloadSizeExceeded, "expected ErrMaxPayloadSizeExceeded, got '%v'", err)
}

func TestSignaler_OnSignalOpts(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer func() {
		_ = sig1.Close()
		_ = sig2.Close()
	}()

	const signal = "onSignalOpts"
	const chanSize = 3
	sig1.AddSignal(signal)
	ln2 := sig2.OnSignalOpts(signal, chanSize)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	data := "test data"

	// Trigger the signals with some test data.
	for i := 0; i < chanSize; i++ {
		checkErr(t, "trigger signal 1", sig1.TriggerSignal(signal, data))
	}

	// Read the signals off of the listener.
	for i := 0; i < chanSize; i++ {
		ctx := <-ln2.C
		var ret string
		checkErr(t, "decoding return data 1", ctx.Decode(&ret))
		assert(t, ret == data, "returned data wrong; expected '%v', got '%v'", data, ret)
	}

	// Now trigger the signal of sig2, which should exceed the configured
	// MaxMessageSize.
	// Trigger the signal with some test data.
	conf2.MaxMessageSize = 1
	err = sig2.TriggerSignal(signal, data)
	assert(t, err == packet.ErrMaxPayloadSizeExceeded, "expected ErrMaxPayloadSizeExceeded, got '%v'", err)
}

func TestSignaler_OnSignalFunc(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer func() {
		_ = sig1.Close()
		_ = sig2.Close()
	}()

	const signal = "onSignalFunc"
	data := "test data"
	sig1.AddSignal(signal)

	errChan := make(chan error)
	retChan := make(chan string)
	ln2 := sig2.OnSignalFunc(signal, func(ctx *signaler.Context) {
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
	err := sig1.TriggerSignal(signal, data)
	checkErr(t, "trigger signal 1", err)

	// Wait for the result.
	select {
	case err = <-errChan:
		checkErr(t, "decoding return data 1", err)
	case ret := <-retChan:
		assert(t, ret == data, "returned data wrong; expected '%v', got '%v'", data, ret)
	}

	// Switch off the listener now.
	ln2.Off()

	// Trigger the signal again.
	err = sig1.TriggerSignal(signal, data)
	checkErr(t, "trigger signal 1", err)

	// Wait for the result.
	select {
	case _ = <-errChan:
		t.Fatal("did not expect an error")
	case _ = <-retChan:
		t.Fatal("did not expect a result")
	case <-time.After(time.Millisecond * 10):
	}
}

func TestSignaler_TriggerSignalError(t *testing.T) {
	t.Parallel()

	sig1, sig2 := testSignaler(nil, nil)
	defer func() {
		_ = sig1.Close()
		_ = sig2.Close()
	}()
	sig1.Ready()
	sig2.Ready()

	// Trigger the a non existent signal.
	err := sig1.TriggerSignal("test", nil)
	assert(t, err == signaler.ErrSignalNotFound, "expected error '%v', got '%v'", signaler.ErrSignalNotFound, err)

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
	err = sig1.TriggerSignal(signal, nil)
	checkErr(t, "trigger inactive signal: %v", err)
}

// convenience
func assert(t *testing.T, condition bool, fmt string, args ...interface{}) {
	if !condition {
		t.Fatalf(fmt, args...)
	}
}

// convenience
func checkErr(t *testing.T, fmt string, err error) {
	if err != nil {
		t.Fatalf(fmt, err)
	}
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
