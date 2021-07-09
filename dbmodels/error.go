package dbmodels

type DBErrCode string

type IDBError interface {
	Error() string
	IsErrorOf(DBErrCode) bool
	ErrCode() DBErrCode
}

const (
	ErrSystemError       DBErrCode = "system_error"
	ErrNoDBRecord        DBErrCode = "no_db_record"
	ErrRecordExists      DBErrCode = "db_record_exists"
	ErrMarshalDataFaield DBErrCode = "failed_to_marshal_data"
)

type dbError struct {
	code DBErrCode
	err  error
}

func (e dbError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e dbError) IsErrorOf(code DBErrCode) bool {
	return e.code == code
}

func (e dbError) ErrCode() DBErrCode {
	return e.code
}

func NewDBError(code DBErrCode, err error) IDBError {
	return dbError{code: code, err: err}
}