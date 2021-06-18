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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"

	"github.com/desertbit/orbit/examples/simple"
	olog "github.com/desertbit/orbit/pkg/hook/log"
	"github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/desertbit/orbit/pkg/transport/quic"
)

func main() {
	// Create the transport for the server.
	// Check pkg/transport for available transports.
	tr, err := quic.NewTransport(&quic.Options{TLSConfig: generateTLSConfig()})
	if err != nil {
		log.Fatalln(err)
	}

	// Create the service. It gets generated from the simple.orbit file.
	// We pass in our custom service handler implementing the server side.
	s, err := simple.NewService(&ServiceHandler{},
		&service.Options{
			ListenAddr: ":1122",
			Transport:  tr,
			Hooks: service.Hooks{
				olog.ServiceHook(), // Pass a logging Hook.
			},
		})
	if err != nil {
		log.Fatalln(err)
	}
	defer s.Close()

	// To start the service and accept incoming connections, we must call Run.
	// This method blocks.
	err = s.Run()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Done, exiting...")
}

// The ServiceHandler type implements the simple.ServiceHandler interface.
type ServiceHandler struct{}

func (sh *ServiceHandler) MyCall(ctx service.Context, arg simple.MyCallArg) (ret simple.MyCallRet, err error) {
	fmt.Printf("MyCall [SUCCESS]: received requiredArg '%s' and optionalArg '%v'\n", arg.RequiredArg, arg.OptionalArg)
	ret.Answer = "Good job"
	return
}

func (sh *ServiceHandler) MyRawStream(ctx service.Context, stream transport.Stream) {
	b := make([]byte, 4)
	n, err := stream.Read(b)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = stream.Write([]byte("pong"))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("MyRawStream [SUCCESS]: received '%s'\n", string(b[:n]))
}

func (sh *ServiceHandler) MyTypedStream(ctx service.Context, stream *simple.MyTypedStreamServiceStream) error {
	arg, err := stream.Read()
	if err != nil {
		return err
	}

	fmt.Printf("MyTypedStream [SUCCESS]: received info about '%s', age '%d'\n", arg.Name, arg.Age)
	return stream.Write(simple.MyTypedStreamRet{Ok: true})
}

// generateTLSConfig generates a bogus tls config that includes a self signed certificate.
// Never use this in a production environment!
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"orbit-simple-example"},
	}
}
