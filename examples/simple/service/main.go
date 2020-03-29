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

	"github.com/desertbit/orbit/examples/simple/hello"
	olog "github.com/desertbit/orbit/pkg/hook/log"
	"github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/desertbit/orbit/pkg/transport/quic"
)

func main() {
	tr, err := quic.NewTransport(&quic.Options{
		TLSConfig: generateTLSConfig(),
	})
	if err != nil {
		log.Fatalln(err)
	}

	s, err := hello.NewService(&ServiceHandler{},
		&service.Options{
			ListenAddr: ":1122",
			Transport:  tr,
			Hooks: service.Hooks{
				olog.ServiceHook(),
			},
		})
	if err != nil {
		log.Fatalln(err)
	}
	defer s.Close()

	err = s.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

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

// #############

type ServiceHandler struct{}

func (s *ServiceHandler) SayHi(ctx service.Context, arg *hello.SayHiArg) (err error) {
	fmt.Println("hi")
	return
}

func (s *ServiceHandler) Test(ctx service.Context) (ret *hello.TestRet, err error) {
	fmt.Println("test")
	ret = &hello.TestRet{
		Name: "horst",
	}
	return
}

func (s *ServiceHandler) Lul(ctx service.Context, stream transport.Stream) {
	panic("implement me")
}

func (s *ServiceHandler) TimeStream(ctx service.Context, arg *hello.InfoReadStream) {
	panic("implement me")
}

func (s *ServiceHandler) ClockTime(ctx service.Context, ret *hello.TimeWriteStream) {
	panic("implement me")
}
