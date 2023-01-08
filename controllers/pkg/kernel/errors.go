package kernel

import "errors"

func ErrorInvalidArgument(msg string) error {
	return errors.New(msg)
}

func ErrorNotFound(msg string) error {
	return errors.New(msg)
}

func ErrorDoesNotExist(msg string) error {
	return errors.New(msg)
}
