package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type ModelErrCode string

const (
	ErrSystemError             ModelErrCode = "system_error"
	ErrUnknownDBError          ModelErrCode = "unknown_db_error"
	ErrWrongVerificationCode   ModelErrCode = "wrong_verification_code"
	ErrVerificationCodeExpired ModelErrCode = "expired_verification_code"
)

type IModelError interface {
	Error() string
	IsErrorOf(ModelErrCode) bool
	ErrCode() ModelErrCode
}

func parseDBError(err dbmodels.IDBError) IModelError {
	if err == nil {
		return nil
	}

	var e error
	e = err

	var code ModelErrCode

	switch err.ErrCode() {
	case dbmodels.ErrMarshalDataFaield:
		code = ErrSystemError

	case dbmodels.ErrSystemError:
		code = ErrSystemError

	default:
		code = ErrUnknownDBError
		e = fmt.Errorf("db code:%s, err:%s", err.ErrCode(), err.Error())
	}

	return newModelError(code, e)
}

type ModelError struct {
	code ModelErrCode
	err  error
}

func (this ModelError) Error() string {
	if this.err == nil {
		return ""
	}
	return this.err.Error()
}

func (this ModelError) IsErrorOf(code ModelErrCode) bool {
	return this.code == code
}

func (this ModelError) ErrCode() ModelErrCode {
	return this.code
}

func newModelError(code ModelErrCode, err error) ModelError {
	return ModelError{code: code, err: err}
}
