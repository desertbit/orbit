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

package quic

import (
	"crypto/tls"
	"errors"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

type Options struct {
	// TODO:
	ListenAddr string

	// TODO:
	DialAddr string

	// TODO:
	Config *quic.Config

	// TODO:
	TLSConfig *tls.Config
}

func DefaultOptions() Options {
	return Options{
		Config: &quic.Config{
			HandshakeTimeout:                      10 * time.Second,
			MaxIdleTimeout:                        30 * time.Second,
			MaxReceiveStreamFlowControlWindow:     6 * 1024 * 1024,  // 6MB
			MaxReceiveConnectionFlowControlWindow: 15 * 1024 * 1024, // 15MB
			KeepAlive:                             true,
		},
	}
}

func (o Options) validate() (err error) {
	if o.ListenAddr == "" && o.DialAddr == "" {
		return errors.New("listen and dial address not set")
	}
	if o.TLSConfig == nil {
		return errors.New("tls config must not be nil")
	}
	return
}
