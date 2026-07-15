package internal

import (
	"errors"
	"strconv"
)

type NotFoundError struct {
	id       int
	itemType string
}

func (e NotFoundError) Error() string {
	return e.itemType + " not found with ID: " + strconv.Itoa(e.id)
}

func (e NotFoundError) Is(target error) bool {
	var notFoundError NotFoundError
	ok := errors.As(target, &notFoundError)
	return ok
}

type ValidationError struct {
	msg string
}

func (e ValidationError) Error() string {
	return e.msg
}

func (e ValidationError) Is(target error) bool {
	var validationErr ValidationError
	ok := errors.As(target, &validationErr)
	return ok
}

func NewValidationErr(message string) ValidationError {
	return ValidationError{
		msg: message,
	}
}

func NewNotFoundError(id int, itemType string) NotFoundError {
	return NotFoundError{id: id, itemType: itemType}
}
