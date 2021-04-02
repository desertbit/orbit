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

package transport

import (
	"context"
	"strings"
	"testing"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/test"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T, clientTr, serverTr Transport) {
	t.Run("listener", testListener(t, clientTr, serverTr))
}

func testListener(t *testing.T, clientTr, serverTr Transport) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// Test bogus address.
		_, err := serverTr.Listen(closer.New(), "blanbitaenr")
		require.Error(t, err)

		// Test closing listener via closer.
		cl := closer.New()
		ln, err := serverTr.Listen(cl, "127.0.0.1:0")
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(ln.Addr().String(), "127.0.0.1:"))
		require.NoError(t, cl.Close())
		_, err = ln.Accept()
		require.Error(t, err)

		// Test a new connection.
		ts := test.NewTester()
		ln, err = serverTr.Listen(closer.New(), "127.0.0.1:0")
		require.NoError(t, err)
		ts.Add(2)
		go ts.Run(func(ctx context.Context, t test.T) {
			conn, err := ln.Accept()
			require.NoError(t, err)
			require.NoError(t, conn.Close())
		})
		go ts.Run(func(ctx context.Context, t test.T) {
			conn, err := clientTr.Dial(closer.New(), ctx, ln.Addr().String())
			require.NoError(t, err)
			require.NoError(t, conn.Close())
		})
		ts.Wait(t)
	}
}

/*func setupTransport(tr Transport) (err error) {
	// Open the listener.
	ln, err := tr.Listen(closer.New(), "localhost")
	if err != nil {
		return
	}

	// Start accepting connections in a new routine.
	go func() {
		defer ln.Close_()

		conn, err := ln.Accept()
		// TODO: handle error from routine.

	}()

	// Wait, until server is up.
	conn, err := tr.Dial(closer.New(), context.Background(), ln.Addr().String())
	if err != nil {
		return
	}
}*/
