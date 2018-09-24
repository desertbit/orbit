package signaler

import "github.com/pkg/errors"

var (
	ErrSignalNotFound       = errors.New("signal not found")
	ErrFilterFuncUndefined = errors.New("could not set filter, filter func is missing")
)
