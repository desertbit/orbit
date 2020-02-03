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

package zerolog

import (
	"net"
	"os"
	"time"

	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog"
)

var _ orbit.Hook = &debug{}

type debug struct {
	log zerolog.Logger
}

func DebugHook() orbit.Hook {
	return DebugHookWithLogger(zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Str("component", "orbit").Logger(),
	)
}

func DebugHookWithLogger(log zerolog.Logger) orbit.Hook {
	return &debug{
		log: log,
	}
}

func (h *debug) OnNewSession(s *orbit.Session, stream net.Conn) error {
	h.log.Debug().
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("new session")
	return nil
}

func (h *debug) OnCall(s *orbit.Session, service, id string) error {
	h.log.Debug().
		Str("service", service).
		Str("id", id).
		Msg("new call")
	return nil
}

func (h *debug) OnCallCompleted(s *orbit.Session, service, id string, err error) {
	h.log.Debug().
		Err(err).
		Str("service", service).
		Str("id", id).
		Msg("call completed")
}

func (h *debug) OnNewStream(s *orbit.Session, service, id string) error {
	h.log.Debug().
		Str("service", service).
		Str("id", id).
		Msg("new stream")
	return nil
}
