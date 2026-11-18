package utils

import (
	"errors"

	"warkop/internal/types"
)

func AsHTTPError(err error) (*types.HTTPError, bool) {
	var he *types.HTTPError
	if errors.As(err, &he) {
		return he, true
	}
	return nil, false
}
