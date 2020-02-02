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

	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/rs/zerolog"
)

var _ orbit.Hook = &hook{}

type hook struct {
	log *zerolog.Logger
}

func Hook(log *zerolog.Logger) orbit.Hook {
	return &hook{
		log: log,
	}
}

func (h *hook) OnNewSession(s *orbit.Session, stream net.Conn) error {
	h.log.Debug().
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("new session")
	return nil
}

func (h *hook) OnNewCall(s *orbit.Session, service, id string) error {
	h.log.Debug().
		Str("service", service).
		Str("id", id).
		Msg("new call")
	return nil
}

func (h *hook) OnCallCompleted(s *orbit.Session, service, id string, err error) {
	h.log.Debug().
		Err(err).
		Str("service", service).
		Str("id", id).
		Msg("call completed")
}

func (h *hook) OnNewStream(s *orbit.Session, service, id string) error {
	h.log.Debug().
		Str("service", service).
		Str("id", id).
		Msg("new stream")
	return nil
}
