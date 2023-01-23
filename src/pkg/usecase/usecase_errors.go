package usecase

type EntityExistsError struct {
	Msg string
}

func (e *EntityExistsError) Error() string {
	return e.Msg
}

func NewEntityExistsError(msg string) *EntityExistsError {
	return &EntityExistsError{
		Msg: msg,
	}
}

type EntityNotFoundError struct {
	Msg string
}

func (e *EntityNotFoundError) Error() string {
	return e.Msg
}

func NewEntityNotFoundError(msg string) *EntityNotFoundError {
	return &EntityNotFoundError{
		Msg: msg,
	}
}

type EntityIncompleteError struct {
	Msg string
}

func (e *EntityIncompleteError) Error() string {
	return e.Msg
}

func NewEntityIncompleteError(msg string) *EntityIncompleteError {
	return &EntityIncompleteError{
		Msg: msg,
	}
}
