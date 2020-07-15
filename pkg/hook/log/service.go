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

package log

import (
	"errors"
	"os"
	"time"

	"github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog"
)

type serviceHook struct {
	log *zerolog.Logger
}

func ServiceHook(logger ...*zerolog.Logger) service.Hook {
	c := &serviceHook{}

	if len(logger) > 0 {
		c.log = logger[0]
	} else {
		l := zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Str("component", "orbit").Logger()
		c.log = &l
	}
	return c
}

func (c *serviceHook) Close() error {
	c.log.Info().Msg("service closed")
	return nil
}

func (c *serviceHook) OnSession(s service.Session, stream transport.Stream) error {
	c.log.Info().
		Str("sessionID", s.ID()).
		Str("localAddr", s.LocalAddr().String()).
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("new session")
	return nil
}

func (c *serviceHook) OnSessionClosed(s service.Session) {
	c.log.Info().
		Str("sessionID", s.ID()).
		Str("localAddr", s.LocalAddr().String()).
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("session closed")
}

func (c *serviceHook) OnCall(ctx service.Context, id string, callKey uint32) error {
	s := ctx.Session()
	c.log.Info().
		Str("callID", id).
		Uint32("callKey", callKey).
		Str("sessionID", s.ID()).
		Str("localAddr", s.LocalAddr().String()).
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("call")
	return nil
}

func (c *serviceHook) OnCallDone(ctx service.Context, id string, callKey uint32, err error) {
	s := ctx.Session()

	if err == nil {
		c.log.Info().
			Str("callID", id).
			Uint32("callKey", callKey).
			Str("sessionID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("call done")
		return
	}

	// Check, if an orbit service error was returned.
	var oErr service.Error
	if errors.As(err, &oErr) {
		c.log.Error().
			Err(err).
			Int("errCode", oErr.Code()).
			Str("errMsg", oErr.Msg()).
			Str("callID", id).
			Uint32("callKey", callKey).
			Str("sessionID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("call failed")
	} else {
		c.log.Error().
			Err(err).
			Str("callID", id).
			Uint32("callKey", callKey).
			Str("sessionID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("call failed")
	}
}

func (c *serviceHook) OnCallCanceled(ctx service.Context, id string, callKey uint32) {
	s := ctx.Session()
	c.log.Info().
		Str("callID", id).
		Uint32("callKey", callKey).
		Str("sessionID", s.ID()).
		Str("localAddr", s.LocalAddr().String()).
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("call canceled")
}

func (c *serviceHook) OnStream(ctx service.Context, id string) error {
	s := ctx.Session()
	c.log.Info().
		Str("streamID", id).
		Str("sessionID", s.ID()).
		Str("localAddr", s.LocalAddr().String()).
		Str("remoteAddr", s.RemoteAddr().String()).
		Msg("stream")
	return nil
}

func (c *serviceHook) OnStreamClosed(ctx service.Context, id string, err error) {
	s := ctx.Session()

	if err == nil {
		c.log.Info().
			Str("streamID", id).
			Str("sessionID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("stream closed")
		return
	}

	// Check, if an orbit client error was returned.
	var oErr service.Error
	if errors.As(err, &oErr) {
		c.log.Error().
			Err(err).
			Int("errCode", oErr.Code()).
			Str("streamID", id).
			Str("sessionID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("stream failed")
	} else {
		c.log.Error().
			Err(err).
			Str("callID", id).
			Str("streamID", s.ID()).
			Str("localAddr", s.LocalAddr().String()).
			Str("remoteAddr", s.RemoteAddr().String()).
			Msg("stream failed")
	}
}
