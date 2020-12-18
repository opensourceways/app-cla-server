package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

var (
	errNoDBRecord  = newDBError(dbmodels.ErrNoDBRecord, fmt.Errorf("no record"))
	errNoChildDoc  = newDBError(dbmodels.ErrNoChildElem, fmt.Errorf("no child record"))
	errRecordExist = newDBError(dbmodels.ErrRecordExists, fmt.Errorf("record exist"))
)

func newDBError(code dbmodels.DBErrCode, err error) *dbmodels.DBError {
	return &dbmodels.DBError{Code: code, Err: err}
}

func systemError(err error) *dbmodels.DBError {
	return newDBError(dbmodels.ErrSystemError, fmt.Errorf("system err:%s", err.Error()))
}
