/*
 * nGin - nLab
 * Copyright (c) 2020 Wahtari GmbH
 */

package auth

import (
	"errors"

	"github.com/desertbit/orbit/pkg/service"
)

// ServiceUserID returns the id of the user stored in the context.
// Returns an error, if its missing.
func ServiceUserID(ctx service.Context) (userID string, err error) {
	userID, ok := ctx.Data(keyUserID).(string)
	if !ok {
		err = errors.New("user id not set in token")
	}
	return
}

// ServiceData returns the data of the user stored in the context.
// Returns an error, if its missing.
func ServiceData(ctx service.Context) (data interface{}, err error) {
	data = ctx.Data(keyData)
	if data == nil {
		err = errors.New("data not set in token")
	}
	return
}
