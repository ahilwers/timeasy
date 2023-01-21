package usecase

type EntityExistsError struct {
	Msg string
}

func (e *EntityExistsError) Error() string {
	return e.Msg
}

type EntityNotFoundError struct {
	Msg string
}

func (e *EntityNotFoundError) Error() string {
	return e.Msg
}
