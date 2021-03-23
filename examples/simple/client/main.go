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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/desertbit/orbit/examples/simple/hello"
	"github.com/desertbit/orbit/examples/simple/world"
	"github.com/desertbit/orbit/pkg/client"
	olog "github.com/desertbit/orbit/pkg/hook/log"
	"github.com/desertbit/orbit/pkg/transport/mux"
	"github.com/desertbit/orbit/pkg/transport/quic"
)

func main() {
	qtr, err := quic.NewTransport(quic.Options{
		DialAddr: "127.0.0.1:1122",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"orbit-simple-example"},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Multiplex transport to allow multiple services.
	mtr, err := mux.New(qtr, mux.DefaultOptions())
	if err != nil {
		return
	}

	c, err := hello.NewClient(client.Options{
		Transport: mtr.Transport("hello"),
		Hooks: client.Hooks{
			olog.ClientHook(),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	wc, err := world.NewClient(client.Options{
		Transport: mtr.Transport("world"),
		Hooks: client.Hooks{
			olog.ClientHook(),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer wc.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ret, err := c.Test(context.Background(), hello.TestArg{S: "testarg"})
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Test: %s, %s\n", ret.Name, ret.Ts.String())
	}()

	ret, err := c.SayHi(context.Background(), hello.SayHiArg{Name: "Wastl", Ts: time.Now()})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("SayHi: %+v\n", ret.Res)

	stream, err := c.ClockTime(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < 3; i++ {
		var arg hello.ClockTimeRet
		arg, err = stream.Read()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("ClockTime: %s\n", arg.Ts.String())
	}

	bi, err := c.Bidirectional(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for i := 0; i < 3; i++ {
		err = bi.Write(hello.BidirectionalArg{Question: "What is the purpose of life?"})
		if err != nil {
			log.Fatalln(err)
		}

		answer, err := bi.Read()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Answer: %s\n", answer.Answer)
	}
	bi.Close()
	wg.Wait()

	err = wc.YetAnotherCall(context.Background(), world.YetAnotherCallArg{S: "Finally done"})
	if err != nil {
		return
	}
}
