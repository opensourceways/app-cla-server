package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type ModelErrCode string

const (
	ErrNoIndividualAndCorpCLA  ModelErrCode = "no_individual_and_corp_cla"
	ErrVerificationCodeExpired ModelErrCode = "expired_verification_code"
	ErrWrongVerificationCode   ModelErrCode = "wrong_verification_code"
	ErrOrgEmailNotExist        ModelErrCode = "org_email_not_exist"
	ErrUnmatchedEmail          ModelErrCode = "unmatched_email"
	ErrNotAnEmail              ModelErrCode = "not_an_email"
	ErrMissingParameter        ModelErrCode = "missing_parameter"
	ErrSystemError             ModelErrCode = "system_error"
	ErrNoLinkOrResign          ModelErrCode = "no_link_or_resign"
	ErrNoLink                  ModelErrCode = "no_link"
	ErrNoCLA                   ModelErrCode = "no_cla"
	ErrNoCorp                  ModelErrCode = "no_corp"
	ErrUnknownDBError          ModelErrCode = "unknown_db_error"
	ErrNoCLAField              ModelErrCode = "no_cla_field"
	ErrManyCLAField            ModelErrCode = "many_cla_field"
	ErrCLAFieldID              ModelErrCode = "invalid_cla_field_id"
	ErrNoOrgSignature          ModelErrCode = "missing_org_signature"
)

func parseDBError(err *dbmodels.DBError) *ModelError {
	if err == nil {
		return nil
	}

	e := err.Err
	var code ModelErrCode

	switch err.Code {
	case dbmodels.ErrMarshalDataFaield:
		code = ErrMissingParameter

	case dbmodels.ErrSystemError:
		code = ErrSystemError

	default:
		code = ErrUnknownDBError
		e = fmt.Errorf("db code:%s, err:%s", err.ErrCode(), err.Error())
	}

	return newModelError(code, e)
}

type ModelError struct {
	Code ModelErrCode
	Err  error
}

func (this *ModelError) Error() string {
	return this.Err.Error()
}

func (this *ModelError) IsErrorOf(code ModelErrCode) bool {
	return this.Code == code
}

func (this *ModelError) ErrCode() string {
	return string(this.Code)
}

func newModelError(code ModelErrCode, err error) *ModelError {
	return &ModelError{Code: code, Err: err}
}
