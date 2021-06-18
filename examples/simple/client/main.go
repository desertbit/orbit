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
	"errors"
	"fmt"
	"log"

	"github.com/desertbit/orbit/examples/simple"
	"github.com/desertbit/orbit/pkg/client"
	olog "github.com/desertbit/orbit/pkg/hook/log"
	"github.com/desertbit/orbit/pkg/transport/quic"
)

func main() {
	// Create the transport for the client.
	// Check pkg/transport for available transports.
	tr, err := quic.NewTransport(&quic.Options{
		TLSConfig: &tls.Config{
			// This config basically disables certificate validation.
			// Never use this in a production environment!
			InsecureSkipVerify: true,
			NextProtos:         []string{"orbit-simple-example"},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Create the client. It gets generated from the simple.orbit file.
	c, err := simple.NewClient(&client.Options{
		Host:      "127.0.0.1:1122",
		Transport: tr,
		Hooks: client.Hooks{
			olog.ClientHook(), // Pass a logging hook into it.
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	// Make a call against the backend.
	// We use a dummy context, but in a production setting you could
	// pass a Context that cancels ongoing requests when the app closes.
	ret, err := c.MyCall(context.Background(), simple.MyCallArg{
		RequiredArg: "I am the required argument", // We must set this argument, otherwise the validation fails.
		OptionalArg: 0,                            // We do not need to set this.
	})
	if err != nil {
		// MyCall may return our custom application error.
		// Check for it by simply using errors.Is.
		if errors.Is(err, simple.ErrCustomErr1) {
			log.Fatalln("MyCall returned our custom error 1")
		}
		log.Fatalln(err)
	}
	fmt.Printf("MyCall [SUCCESS]: returned %s\n", ret.Answer)

	// Open our typed stream.
	// It allows us to send specific types on a stream connection.
	// When we are done with it, we must close the stream.
	stream, err := c.MyTypedStream(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	defer stream.Close()

	// Send arguments over the stream.
	err = stream.Write(simple.PersonInfo{
		Name:        "Marcus",
		Age:         25,
		Locale:      "en_US",
		Address:     "Dream Boulevard 25",
		VehicleType: simple.Car,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Read the answer from the service.
	sRet, err := stream.Read()
	if err != nil {
		if errors.Is(err, simple.ErrCustomErr1) {
			log.Fatalln("MyTypedStream.Read returned our custom error 1")
		} else if errors.Is(err, simple.ErrCustomErr2) {
			log.Fatalln("MyTypedStream.Read returned our custom error 2")
		}
		log.Fatalln(err)
	}
	fmt.Printf("MyTypedStream [SUCCESS]: returned %v\n", sRet.Ok)

	// Open our raw stream.
	// You could do basically anything with it.
	// We just send a ping pong.
	rawStream, err := c.MyRawStream(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	_, err = rawStream.Write([]byte("ping"))
	if err != nil {
		log.Fatalln(err)
	}

	b := make([]byte, 4)
	n, err := rawStream.Read(b)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("MyRawStream [SUCCESS]: returned %s\n", string(b[:n]))
	fmt.Println("Done, exiting...")
}
