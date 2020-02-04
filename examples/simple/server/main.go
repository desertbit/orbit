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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/hook/auth"
	"github.com/desertbit/orbit/pkg/hook/zerolog"
	"github.com/desertbit/orbit/pkg/net/quic"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog/log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal().Err(err).Msg("server run")
	}
}

func run() (err error) {
	cl := closer.New()
	defer cl.Close_()

	// Create the listener.
	/*ln, err := yamux.NewTCPListener("127.0.0.1:6789", nil)
	if err != nil {
		return
	}*/

	ln, err := quic.NewUDPListenerWithCloser("127.0.0.1:6789", generateTLSConfig(), quic.DefaultConfig(), cl.CloserTwoWay())
	if err != nil {
		return
	}

	// Create the config.
	cf := &orbit.ServerConfig{Config: orbit.DefaultConfig()}
	cf.PrintPanicStackTraces = true
	cf.SendErrToCaller = true

	// Create the user hash func.
	uhf := func(username string) (hash []byte, ok bool) {
		if username == "marc" {
			hash, _ = auth.Hash("test")
			ok = true
		}
		return
	}

	// Create the orbit server.
	s := orbit.NewServer(ln, cf, &Server{}, auth.ServerHook(uhf), zerolog.DebugHookWithLogger(*cf.Log))

	go func() {
		err := s.Listen()
		if err != nil && !s.IsClosing() {
			log.Error().Err(err).Msg("server listen")
		}
	}()

	log.Info().Msg("listening...")

	// Wait until closed.
	go func() {
		// Wait for the signal.
		sigchan := make(chan os.Signal, 3)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigchan

		log.Info().Msg("Exiting...")
		cl.Close_()
	}()

	<-cl.ClosedChan()
	return
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
