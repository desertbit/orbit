package signaler

import "github.com/pkg/errors"

var (
	// ErrSignalNotFound is an error indicating that a signal could not be found.
	ErrSignalNotFound = errors.New("signal not found")
	// ErrFilterFuncUndefined is an error indicating that a signal is missing its
	// filter func.
	ErrFilterFuncUndefined = errors.New("could not set filter, filter func is missing")
)
