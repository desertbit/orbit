/*
 * nGin - nLab
 * Copyright (c) 2020 Wahtari GmbH
 */

package auth

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/service"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	keyHeaderToken = "auth-token"

	// Data keys.
	keyUserID = "auth-user-id"
	keyData   = "auth-data"
)

var ErrAuthFailed = errors.New("auth failed")

type ServiceHookHandler interface {
	Authenticate(tk *Token, callID string) (err error)
}

type NoAuth map[string]struct{}

type serviceHook struct {
	cs      crypto.Signer
	csOpts  crypto.SignerOpts
	sigSize int
	noAuth  NoAuth
	handler ServiceHookHandler
	codec   codec.Codec
}

func ServiceHook(
	cs crypto.Signer,
	csOpts crypto.SignerOpts,
	sigSize int,
	na NoAuth,
	h ServiceHookHandler,
	cc codec.Codec,
) service.Hook {
	return &serviceHook{
		cs:      cs,
		csOpts:  csOpts,
		sigSize: sigSize,
		noAuth:  na,
		handler: h,
		codec:   cc,
	}
}

// Implements the service.Hook interface.
func (serviceHook) Close() error { return nil }

// Implements the service.Hook interface.
func (serviceHook) OnSession(s service.Session, stream transport.Stream) error { return nil }

// Implements the service.Hook interface.
func (serviceHook) OnSessionClosed(s service.Session) {}

// Implements the service.Hook interface.
func (s serviceHook) OnCall(ctx service.Context, id string, callKey uint32) error {
	// Verify the token.
	err := s.verifyToken(ctx, id)
	if err != nil {
		return fmt.Errorf("auth.serviceHook.OnCall: %w", errAuthFailed(err))
	}

	return nil
}

// Implements the service.Hook interface.
func (serviceHook) OnCallDone(ctx service.Context, id string, callKey uint32, err error) {}

// Implements the service.Hook interface.
func (serviceHook) OnCallCanceled(ctx service.Context, id string, callKey uint32) {}

// Implements the service.Hook interface.
func (s serviceHook) OnStream(ctx service.Context, id string) error {
	// Verify the token.
	err := s.verifyToken(ctx, id)
	if err != nil {
		return fmt.Errorf("auth.serviceHook.OnStream: %w", errAuthFailed(err))
	}

	return nil
}

// Implements the service.Hook interface.
func (s serviceHook) OnStreamClosed(ctx service.Context, id string, err error) {}

//###############//
//### Private ###//
//###############//

// verifyToken extracts the token data from the context, validates it
// against the defined access rules and sets upon success the payload data of the
// token into the data map of the context.
func (s serviceHook) verifyToken(ctx service.Context, id string) error {
	// Check, if the call must be authenticated.
	if _, ok := s.noAuth[id]; ok {
		return nil
	}

	// Retrieve the token from the headers.
	tokenData := ctx.Header(keyHeaderToken)
	tkdLen := len(tokenData)
	if tkdLen == 0 {
		return errors.New("token data is missing")
	} else if tkdLen <= s.sigSize {
		return errors.New("token data size invalid")
	}

	// Split the data into token and signature.
	tkLen := tkdLen - s.sigSize
	token, tokenSig := tokenData[:tkLen], tokenData[tkLen:]

	// Sign the token and then check, if the signatures match.
	sig, err := s.cs.Sign(rand.Reader, token, s.csOpts)
	if err != nil {
		return err
	} else if !bytes.Equal(sig, tokenSig) {
		return errors.New("token signature is invalid")
	}

	// Decode and validate the token.
	tk := &Token{}
	err = s.codec.Decode(token, tk)
	if err != nil {
		return err
	}
	err = tk.Valid(time.Now())
	if err != nil {
		return fmt.Errorf("token invalid: %w", err)
	}

	// Token has been successfully verified.
	// Add the userID and groups to the context data.
	ctx.SetData(keyUserID, tk.UserID)
	ctx.SetData(keyData, tk.Data)

	return nil
}

func errAuthFailed(err error) error {
	return fmt.Errorf("%w: %v", ErrAuthFailed, err)
}
