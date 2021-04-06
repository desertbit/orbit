/*
 * Guru - nLab
 * Copyright (c) 2020 Wahtari GmbH
 */

package auth_test

import (
	"testing"
	"time"

	"github.com/desertbit/orbit/pkg/hook/auth"
	"github.com/stretchr/testify/require"
)

func TestToken_Valid(t *testing.T) {
	t.Parallel()

	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	zeroISS := time.Time{}
	zeroEXP := time.Time{}
	validISS := t1.Add(-5 * time.Minute)
	validEXP := t1.Add(5 * time.Minute)

	cases := []struct {
		userID                       string
		issuedOn, expiresOn, valTime time.Time
		valid                        bool
	}{
		{userID: "", issuedOn: validISS, expiresOn: validEXP, valTime: t1, valid: false}, // 0
		{userID: "test", issuedOn: zeroISS, expiresOn: validEXP, valTime: t1, valid: false},
		{userID: "test", issuedOn: validISS, expiresOn: zeroEXP, valTime: t1, valid: false},
		{userID: "test", issuedOn: zeroISS, expiresOn: zeroEXP, valTime: t1, valid: false},
		{userID: "test", issuedOn: t1.Add(time.Nanosecond), expiresOn: validEXP, valTime: t1, valid: false},
		{userID: "test", issuedOn: validISS, expiresOn: t1.Add(-time.Nanosecond), valTime: t1, valid: false}, // 5
		{userID: "test", issuedOn: validISS, expiresOn: validEXP, valTime: t1, valid: true},
	}

	for i, c := range cases {
		tk := &auth.Token{UserID: c.userID, IssuedOn: c.issuedOn, ExpiresOn: c.expiresOn}

		if c.valid {
			require.NoError(t, tk.Valid(c.valTime), "case %d", i)
		} else {
			require.Error(t, tk.Valid(c.valTime), "case %d", i)
		}
	}
}
