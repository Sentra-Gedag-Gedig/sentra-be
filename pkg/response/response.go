package response

import (
	"errors"
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Is(target error) bool {
	var t *Error
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Err.Error() == t.Err.Error()
}

func NewError(code int, err string) error {
	return &Error{code, errors.New(err)}
}
