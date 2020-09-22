package dbmodels

type ErrHasSigned struct {
	Err error
}

func (this ErrHasSigned) Error() string {
	return this.Err.Error()
}

func IsHasSigned(err error) bool {
	_, ok := err.(ErrHasSigned)
	return ok
}
