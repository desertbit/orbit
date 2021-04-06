/*
 * nGin - nLab
 * Copyright (c) 2020 Wahtari GmbH
 */

package auth

import (
	"github.com/desertbit/orbit/pkg/client"
	"github.com/desertbit/orbit/pkg/transport"
)

type ClientHookHandler interface {
	Token() []byte
}

type clientHook struct {
	h ClientHookHandler
}

func ClientHook(h ClientHookHandler) *clientHook {
	return &clientHook{h: h}
}

// Implements the client.Hook interface.
func (c *clientHook) Close() error { return nil }

// Implements the client.Hook interface.
func (c *clientHook) OnSession(s client.Session, stream transport.Stream) error { return nil }

// Implements the client.Hook interface.
func (c *clientHook) OnSessionClosed(s client.Session) {}

// Implements the client.Hook interface.
func (c *clientHook) OnCall(ctx client.Context, id string, callKey uint32) error {
	// Get the auth token from the handler.
	tk := c.h.Token()
	if tk == nil {
		return nil
	}

	// Add the auth token to the header.
	ctx.SetHeader(keyHeaderToken, tk)
	return nil
}

// Implements the client.Hook interface.
func (c *clientHook) OnCallDone(ctx client.Context, id string, callKey uint32, err error) {}

// Implements the client.Hook interface.
func (c *clientHook) OnCallCanceled(ctx client.Context, id string, callKey uint32) {}

// Implements the client.Hook interface.
func (c *clientHook) OnStream(ctx client.Context, id string) error {
	// Get the auth token from the handler.
	tk := c.h.Token()
	if tk == nil {
		return nil
	}

	// Add the token to the header.
	ctx.SetHeader(keyHeaderToken, tk)
	return nil
}

// Implements the client.Hook interface.
func (c *clientHook) OnStreamClosed(ctx client.Context, id string) {}
