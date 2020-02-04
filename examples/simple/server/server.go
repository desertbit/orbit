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
	"sync"
	"time"

	"github.com/desertbit/orbit/examples/simple/hello"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog/log"
)

type Server struct {
	os *orbit.Server

	sessionsMx sync.RWMutex
	sessions   map[string]*Session
}

func (s *Server) InitServer(os *orbit.Server) {
	s.os = os
}

func (s *Server) InitSession(osn *orbit.Session) {
	sn := &Session{}
	sn.HelloServerCaller = hello.RegisterHelloServer(osn, sn)
	osn.SetValue("HelloServer", sn)

	go func() {
		time.Sleep(time.Second)
		err := sn.SayHi(context.Background(), osn, &hello.SayHiArgs{Name: "Roland"})
		if err != nil {
			log.Error().Err(err).Msg("say hi to client")
		}
	}()
}

func (s *Server) Session(id string) (sn *Session) {
	osn := s.os.Session(id)
	if osn == nil {
		return
	}

	sn, _ = osn.Value("HelloServer").(*Session)
	return
}
