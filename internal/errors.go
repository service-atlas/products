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

type ValidationError struct {
	msg string
}

func (e ValidationError) Error() string {
	return e.msg
}

func NewValidationErr(message string) ValidationError {
	return ValidationError{
		msg: message,
	}
}

func NewNotFoundError(id int, itemType string) NotFoundError {
	return NotFoundError{id: id, itemType: itemType}
}

func IsNotFoundError(err error) bool {
	_, ok := errors.AsType[NotFoundError](err)
	return ok
}

func IsValidationError(err error) bool {
	_, ok := errors.AsType[ValidationError](err)
	return ok
}
