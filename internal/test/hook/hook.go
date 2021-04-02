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

package hook

import (
	"github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport"
)

type ServiceHook struct {
	CloseFunc           func() error
	OnListeningFunc     func(listenAddr string)
	OnSessionFunc       func(s service.Session, stream transport.Stream) error
	OnSessionClosedFunc func(s service.Session)
	OnCallFunc          func(ctx service.Context, id string, callKey uint32) error
	OnCallDoneFunc      func(ctx service.Context, id string, callKey uint32, err error)
	OnCallCanceledFunc  func(ctx service.Context, id string, callKey uint32)
	OnStreamFunc        func(ctx service.Context, id string) error
	OnStreamClosedFunc  func(ctx service.Context, id string, err error)
}

func (h ServiceHook) Close() error {
	if h.CloseFunc != nil {
		return h.CloseFunc()
	}
	return nil
}

func (h ServiceHook) OnListening(listenAddr string) {
	if h.OnListeningFunc != nil {
		h.OnListeningFunc(listenAddr)
	}
}

func (h ServiceHook) OnSession(s service.Session, stream transport.Stream) error {
	if h.OnSessionFunc != nil {
		return h.OnSessionFunc(s, stream)
	}
	return nil
}

func (h ServiceHook) OnSessionClosed(s service.Session) {
	if h.OnSessionClosedFunc != nil {
		h.OnSessionClosedFunc(s)
	}
}

func (h ServiceHook) OnCall(ctx service.Context, id string, callKey uint32) error {
	if h.OnCallFunc != nil {
		return h.OnCallFunc(ctx, id, callKey)
	}
	return nil
}

func (h ServiceHook) OnCallDone(ctx service.Context, id string, callKey uint32, err error) {
	if h.OnCallDoneFunc != nil {
		h.OnCallDoneFunc(ctx, id, callKey, err)
	}
}

func (h ServiceHook) OnCallCanceled(ctx service.Context, id string, callKey uint32) {
	if h.OnCallCanceledFunc != nil {
		h.OnCallCanceledFunc(ctx, id, callKey)
	}
}

func (h ServiceHook) OnStream(ctx service.Context, id string) error {
	if h.OnStreamFunc != nil {
		return h.OnStreamFunc(ctx, id)
	}
	return nil
}

func (h ServiceHook) OnStreamClosed(ctx service.Context, id string, err error) {
	if h.OnStreamClosedFunc != nil {
		h.OnStreamClosedFunc(ctx, id, err)
	}
}
