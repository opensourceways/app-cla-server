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
	ErrNotFound          DBErrCode = "not_found"
	ErrRecordExists      DBErrCode = "db_record_exists"
	ErrMarshalDataFaield DBErrCode = "failed_to_marshal_data"
)
