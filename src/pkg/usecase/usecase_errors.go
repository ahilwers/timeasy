package usecase

type EntityExistsError struct {
	Msg string
}

func (e *EntityExistsError) Error() string {
	return e.Msg
}
