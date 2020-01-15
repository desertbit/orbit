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
	"os"
	"os/signal"
	"syscall"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/net/yamux"
	"github.com/desertbit/orbit/pkg/orbit"
	yamux2 "github.com/hashicorp/yamux"
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

	ln, err := yamux.NewTCPListener("127.0.0.1:6789", yamux2.DefaultConfig())
	if err != nil {
		return
	}

	orbServ := orbit.NewServer(
		cl.CloserTwoWay(),
		ln,
		&orbit.ServerConfig{
			Config: &orbit.Config{PrintPanicStackTraces: true},
		},
	)

	s := NewServer(orbServ)
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
