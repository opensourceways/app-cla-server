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

func IsDBError(err error) (DBError, bool) {
	e, ok := err.(DBError)
	return e, ok
}

func IsHasSigned(err DBError) bool {
	return err.ErrCode == ErrHasSigned
}

func IsInvalidParameter(err DBError) bool {
	return err.ErrCode == ErrInvalidParameter
}
