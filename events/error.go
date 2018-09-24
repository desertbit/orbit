package events

import "github.com/pkg/errors"

var (
	ErrEventNotFound = errors.New("event not found")

	errFilterFuncUndefined = errors.New("could not set filter, filter func is missing")
)
