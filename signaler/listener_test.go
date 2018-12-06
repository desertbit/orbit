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
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/signaler"
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
	checkErr(t, "trigger signal", sig1.TriggerSignal(signal, data))

	// Wait for the result.
	select {
	case _ = <-errChan:
		t.Fatal("did not expect an error")
	case _ = <-retChan:
		t.Fatal("did not expect a result")
	case <-time.After(time.Millisecond * 10):
	}
}
