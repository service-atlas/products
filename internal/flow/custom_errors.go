package flow

type DependencyValidationError struct {
}

func (e DependencyValidationError) Error() string {
	return "required data dependency not found"
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}
