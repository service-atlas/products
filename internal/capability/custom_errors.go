package capability

import "errors"

type NotFoundError struct {
	Msg string
}

func (e NotFoundError) Error() string {
	return e.Msg
}

func (e NotFoundError) Is(target error) bool {
	var t NotFoundError
	return errors.As(target, &t)
}
