package dbmodels

type DBError struct {
	ErrCode string
	Err     error
}

func (this DBError) Error() string {
	return this.Err.Error()
}

func IsDBError(err error) (DBError, bool) {
	e, ok := err.(DBError)
	return e, ok
}
