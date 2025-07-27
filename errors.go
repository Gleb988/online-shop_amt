package amt

type ErrNackType struct {
	Message string
}

func (e ErrNackType) Error() string {
	return e.Message
}

func (e ErrNackType) Is(target error) bool {
	return e.Message == target.Error()
}
