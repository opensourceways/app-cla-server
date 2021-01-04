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
	ErrUnmatchedEmail          ModelErrCode = "unmatched_email"
	ErrNotAnEmail              ModelErrCode = "not_an_email"
	ErrNoLink                  ModelErrCode = "no_link"
	ErrNoLinkOrResign          ModelErrCode = "no_link_or_resign"
	ErrNoLinkOrUnsigned        ModelErrCode = "no_link_or_unsigned"
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

type modelError struct {
	code ModelErrCode
	err  error
}

func (this modelError) Error() string {
	if this.err == nil {
		return ""
	}
	return this.err.Error()
}

func (this modelError) IsErrorOf(code ModelErrCode) bool {
	return this.code == code
}

func (this modelError) ErrCode() ModelErrCode {
	return this.code
}

func newModelError(code ModelErrCode, err error) IModelError {
	return modelError{code: code, err: err}
}
