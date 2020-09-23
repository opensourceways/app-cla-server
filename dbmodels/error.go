package dbmodels

const (
	ErrorStart = iota
	ErrInvalidParameter
	ErrHasSigned
)

type DBError struct {
	ErrCode int
	Err     error
}

func (this DBError) Error() string {
	return this.Err.Error()
}

func IsHasSigned(err error) bool {
	e, ok := err.(DBError)
	return ok && e.ErrCode == ErrHasSigned
}

func IsInvalidParameter(err error) bool {
	e, ok := err.(DBError)
	return ok && e.ErrCode == ErrInvalidParameter
}
