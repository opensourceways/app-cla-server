package dbmodels

const (
	ErrorStart = iota
	ErrInvalidParameter
	ErrHasSigned
	ErrHasNotSigned
)

type DBError struct {
	ErrCode int
	Err     error
}

func (this DBError) Error() string {
	return this.Err.Error()
}

func IsDBError(err error) (DBError, bool) {
	e, ok := err.(DBError)
	return e, ok
}
