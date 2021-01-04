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

type DBErrCode string

type IDBError interface {
	Error() string
	IsErrorOf(DBErrCode) bool
	ErrCode() DBErrCode
}

const (
	ErrMarshalDataFaield DBErrCode = "failed_to_marshal_data"
	ErrNoDBRecord        DBErrCode = "no_db_record"
	ErrSystemError       DBErrCode = "system_error"
)
