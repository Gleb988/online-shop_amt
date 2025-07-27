package models

type ErrNack struct {
	Message string
}

func (e ErrNack) Error() string {
	return e.Message
}

func (e ErrNack) Is(target error) bool {
	return e.Message == target.Error()
}
