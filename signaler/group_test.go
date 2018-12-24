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
	"testing"
	"time"

	"github.com/desertbit/orbit/signaler"
	"github.com/stretchr/testify/require"
)

// TestGroup tests the complete group file.
func TestGroup(t *testing.T) {
	sigServer1, sigClient2 := testSignaler(nil, nil)
	sigServer3, sigClient4 := testSignaler(nil, nil)
	defer closeSignalers(sigServer1, sigClient2, sigServer3, sigClient4)

	const signal = "test"
	data := "testData"

	resChan := make(chan string, 2)
	errChan := make(chan error, 2)

	handlerF := func(ctx *signaler.Context) {
		var data string
		err := ctx.Decode(&data)
		if err != nil {
			errChan <- err
			return
		}

		resChan <- data
	}

	sigServer1.AddSignal(signal)
	sigServer3.AddSignal(signal)

	sigClient2.OnSignalFunc(signal, handlerF)
	sigClient4.OnSignalFunc(signal, handlerF)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Create the group.
	group := signaler.NewGroup()
	group.Add(sigServer1, sigServer3)

	// Test an empty Add.
	group.Add()
	// Test a nil Add.
	group.Add(nil)

	// Normal test, both clients should receive the signal.
	err := group.Trigger(signal, data)
	require.NoError(t, err)

	timeout := 10 * time.Millisecond

	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			t.Fatalf("error 1: %v", err)
		case res := <-resChan:
			require.Equal(t, data, res)
		case <-time.After(timeout):
			t.Fatal("timeout 1")
		}
	}

	// Exclude sigServer3 from trigger, only sigClient2 should receive the signal.
	err = group.Trigger(signal, data, sigServer3)
	require.NoError(t, err)

	// Get result from sigClient2.
	select {
	case err := <-errChan:
		t.Fatalf("error 2: %v", err)
	case res := <-resChan:
		require.Equal(t, data, res)
	case <-time.After(timeout):
		t.Fatal("timeout 2")
	}

	// Ensure sigClient4 does not send something back.
	select {
	case err := <-errChan:
		t.Fatalf("error 3: %v", err)
	case _ = <-resChan:
		t.Fatal("did not expect result 3")
	case <-time.After(timeout):
	}

	// Exclude both server signals, no client should receive the signal.
	err = group.Trigger(signal, data, sigServer1, sigServer3)
	require.NoError(t, err)

	select {
	case err := <-errChan:
		t.Fatalf("error 4: %v", err)
	case _ = <-resChan:
		t.Fatal("did not expect result 4")
	case <-time.After(timeout):
	}

	// Trigger a signal that does not exist. This is allowed and should
	// not produce an error.
	err = group.Trigger("blabla", data)
	require.NoError(t, err)

	// Close one of the signalers and try again. This should now produce
	// an error.
	err = sigServer1.Close()
	require.NoError(t, err)

	err = group.Trigger(signal, "test")
	require.Error(t, err, "expecting error for closed signaler")
	require.NotEqual(t, signaler.ErrSignalNotFound, err)
}
