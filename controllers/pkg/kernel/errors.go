package kernel

import "errors"

func ErrorInvalidArgument(msg string) error {
	return errors.New(msg)
}
