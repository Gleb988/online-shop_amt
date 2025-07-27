package amt

// особый тип ошибок для того, чтобы не возвращать сообщение в очередь
type ErrNack struct {
	Message string
}

func (e ErrNack) Error() string {
	return e.Message
}

func (e ErrNack) Is(target error) bool {
	return e.Message == target.Error()
}

func NewErrNack(message string) error {
	return ErrNack{Message: message}
}
