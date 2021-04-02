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

package call_test

import (
	"context"
	"testing"
	"time"

	"github.com/desertbit/orbit/internal/test"
	"github.com/desertbit/orbit/internal/test/call"
	"github.com/desertbit/orbit/internal/test/hook"
	"github.com/desertbit/orbit/pkg/client"
	oservice "github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport/yamux"
	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	t.Run("callData", testCallData)
	t.Run("callMaxSizes", testCallMaxSizes)
}

func testCallData(t *testing.T) {
	t.Parallel()
	ts := test.NewTester()

	// Define transport.
	tr, err := yamux.NewTransport(&yamux.Options{})
	require.NoError(t, err)

	// Define Hook.
	addrChan := make(chan string, 1)
	sh := hook.ServiceHook{
		OnListeningFunc: func(listenAddr string) { addrChan <- listenAddr },
	}

	// Define service handler.
	handler := &serviceHandler{}

	// Create the service.
	srv, err := call.NewService(handler, &oservice.Options{
		ListenAddr: "localhost:50000",
		Transport:  tr,
		Hooks:      oservice.Hooks{sh},
	})
	require.NoError(t, err)
	ts.Add(1)
	go ts.Run(func(ctx context.Context, t test.T) {
		require.NoError(t, srv.Run())
	})

	// Create the client.
	listenAddr := <-addrChan
	client, err := call.NewClient(&client.Options{
		Host:      listenAddr,
		Transport: tr,
	})
	require.NoError(t, err)

	// Send a request with no data.
	err = client.NoData(context.Background())
	require.NoError(t, err)

	// Send a request with arguments.
	data := []byte("this is some really nice test data")
	ts.Add(1)
	handler.ArgDataFunc = func(ctx oservice.Context, arg call.ArgDataArg) error {
		ts.Run(func(_ context.Context, t test.T) {
			require.Equal(t, data, arg.Data)
		})
		return nil
	}
	err = client.ArgData(context.Background(), call.ArgDataArg{Data: data})
	require.NoError(t, err)

	// Send a request with return data.
	data2 := map[string][]int{"test": {5, 6, 8, 1}, "hello": {669}}
	handler.RetDataFunc = func(ctx oservice.Context) (ret call.RetDataRet, err error) {
		ret.Data = data2
		return
	}
	ret, err := client.RetData(context.Background())
	require.NoError(t, err)
	require.Equal(t, data2, ret.Data)

	// Send a request with both args and ret.
	ts.Add(1)
	argData := 85.5
	retData1 := time.Unix(8468455484, 99548)
	retData2 := []map[string][]bool{
		{"ha": {true, true, false}},
		{"second": nil},
	}
	handler.ArgRetDataFunc = func(ctx oservice.Context, arg call.ArgRetDataArg) (ret call.ArgRetDataRet, err error) {
		ts.Run(func(_ context.Context, t test.T) {
			require.Exactly(t, argData, arg.Data)
		})
		ret.Data1 = retData1
		ret.Data2 = retData2
		return
	}
	ret2, err := client.ArgRetData(context.Background(), call.ArgRetDataArg{Data: argData})
	require.NoError(t, err)
	require.True(t, retData1.Equal(ret2.Data1))
	require.Equal(t, retData2, ret2.Data2)

	require.NoError(t, srv.Close())
	ts.Wait(t)
}

func testCallMaxSizes(t *testing.T) {
	t.Parallel()
}

//#######################//
//### Service Handler ###//
//#######################//

type serviceHandler struct {
	NoDataFunc     func(ctx oservice.Context) error
	ArgDataFunc    func(ctx oservice.Context, arg call.ArgDataArg) error
	RetDataFunc    func(ctx oservice.Context) (ret call.RetDataRet, err error)
	ArgRetDataFunc func(ctx oservice.Context, arg call.ArgRetDataArg) (ret call.ArgRetDataRet, err error)
}

func (h *serviceHandler) NoData(ctx oservice.Context) (err error) {
	if h.NoDataFunc != nil {
		err = h.NoDataFunc(ctx)
	}
	return
}

func (h *serviceHandler) ArgData(ctx oservice.Context, arg call.ArgDataArg) (err error) {
	if h.ArgDataFunc != nil {
		err = h.ArgDataFunc(ctx, arg)
	}
	return
}

func (h *serviceHandler) RetData(ctx oservice.Context) (ret call.RetDataRet, err error) {
	if h.RetDataFunc != nil {
		ret, err = h.RetDataFunc(ctx)
	}
	return
}

func (h *serviceHandler) ArgRetData(ctx oservice.Context, arg call.ArgRetDataArg) (ret call.ArgRetDataRet, err error) {
	if h.ArgRetDataFunc != nil {
		ret, err = h.ArgRetDataFunc(ctx, arg)
	}
	return
}
