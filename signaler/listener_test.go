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
	"log"
	"testing"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/signaler"
	"github.com/stretchr/testify/require"
)

func TestListener_OffAndOffChan(t *testing.T) {
	t.Parallel()

	buffer1 := bytes.Buffer{}
	buffer2 := bytes.Buffer{}
	sig1, sig2 := testSignaler(
		&control.Config{Logger: log.New(&buffer1, "", 0)},
		&control.Config{Logger: log.New(&buffer2, "", 0)},
	)
	defer closeSignalers(sig1, sig2)

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

	// Switch off the listener now.
	go func() {
		ln2.Off()
	}()

	// Check, if the off chan has received a signal.
	select {
	case <-ln2.OffChan():
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout when waiting for off chan")
	}

	// Trigger the signal.
	require.NoError(t, sig1.TriggerSignal(signal, data))

	// Wait for the result.
	select {
	case _ = <-errChan:
		t.Fatal("did not expect an error")
	case _ = <-retChan:
		t.Fatal("did not expect a result")
	case <-time.After(time.Millisecond * 10):
	}
}
