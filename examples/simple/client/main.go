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

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/examples/simple/hello"
	"github.com/desertbit/orbit/pkg/hook/auth"
	"github.com/desertbit/orbit/pkg/net/quic"
	"github.com/desertbit/orbit/pkg/orbit"
	quic2 "github.com/lucas-clemente/quic-go"
	"github.com/rs/zerolog/log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("client run")
	}
}

func run() (err error) {
	cl := closer.New()
	defer cl.Close_()

	// Create the tcp connection.
	/*conn, err := yamux.NewTCPConn("127.0.0.1:6789", nil)
	if err != nil {
		return
	}*/

	conn, err := quic.NewUDPConnWithCloser("127.0.0.1:6789", &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"orbit-simple-example"},
	}, quic.DefaultConfig(), cl.CloserTwoWay())
	if err != nil {
		err = fmt.Errorf("new udp conn with closer: %w", err)
		return
	}

	// Create the config.
	cf := orbit.DefaultConfig()
	cf.PrintPanicStackTraces = true
	cf.SendErrToCaller = true

	// Create our client.
	c := &Client{}

	// Create the orbit client.
	_, err = orbit.NewClient(conn, cf, c, auth.ClientHook("marc", "test"))
	if err != nil {
		return
	}

	// Make example calls.
	for i := 0; i < 1000; i++ {
		err = c.SayHi(context.Background(), &hello.SayHiArgs{Name: "Marc"})
		if err != nil {
			var t net.Error
			if errors.As(err, &t) {
				println("main say hi NET ERROR", t.Timeout(), t.Temporary(), t.Error())
			}
			var s quic2.StreamError
			if errors.As(err, &s) {
				println("main say hi STREAM ERROR", s.ErrorCode(), s.Canceled(), s.Error())
			}
			err = fmt.Errorf("say hi: %w", err)
			return
		}
	}
	println("ok")

	/*ret, err := c.ClockTime(context.Background())
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		var t time.Time
		t, err = ret.Read()
		if err != nil {
			return
		}

		log.Debug().Time("time", t).Msg("server clock time")
	}
	return ret.Close()*/
	return
}
