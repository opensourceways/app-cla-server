package dbmodels

type DBErrCode string

type DBError struct {
	Code DBErrCode
	Err  error
}

func (this DBError) Error() string {
	return this.Err.Error()
}

func (this DBError) IsErrorOf(code DBErrCode) bool {
	return this.Code == code
}

func (this DBError) ErrCode() string {
	return string(this.Code)
}

const (
	ErrMarshalDataFaield DBErrCode = "failed_to_marshal_data"
	ErrNoDBRecord        DBErrCode = "no_db_record"
	ErrSystemError       DBErrCode = "system_error"
	ErrRecordExists      DBErrCode = "db_record_exists"
)
