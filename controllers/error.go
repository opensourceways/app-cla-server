package controllers

import "github.com/opensourceways/app-cla-server/models"

const (
	errNotSameCorp           = "not_same_corp"
	errHasSigned             = "has_signed"
	errHasNotSigned          = "has_not_signed"
	errSystemError           = "system_error"
	errUnmatchedCLA          = "unmatched_cla"
	errUnknownLink           = "unknown_link"
	ErrNoCorpEmployeeManager = "no_corp_employee_manager"
)

func parseModelError(err *models.ModelError) *failedResult {
	if err == nil {
		return nil
	}

	sc := 400
	code := ""
	switch err.Code {
	case models.ErrUnknownDBError:
		sc = 500
		code = errSystemError

	case models.ErrSystemError:
		sc = 500
		code = errSystemError

	default:
		code = err.ErrCode()
	}

	return newFailedResult(sc, code, err)
}
