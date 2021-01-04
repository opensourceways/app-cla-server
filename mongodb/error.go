package mongodb

import "github.com/opensourceways/app-cla-server/dbmodels"

type dbError struct {
	code dbmodels.DBErrCode
	err  error
}

func (this dbError) Error() string {
	if this.err == nil {
		return ""
	}
	return this.err.Error()
}

func (this dbError) IsErrorOf(code dbmodels.DBErrCode) bool {
	return this.code == code
}

func (this dbError) ErrCode() dbmodels.DBErrCode {
	return this.code
}

func newDBError(code dbmodels.DBErrCode, err error) dbmodels.IDBError {
	return dbError{code: code, err: err}
}

func newSystemError(err error) dbmodels.IDBError {
	return dbError{code: dbmodels.ErrSystemError, err: err}
}
