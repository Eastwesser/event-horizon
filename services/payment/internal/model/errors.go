package model

import "errors"

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
