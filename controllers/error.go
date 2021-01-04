package controllers

import "github.com/opensourceways/app-cla-server/models"

const (
	errSystemError      = "system_error"
	errMissingToken     = "missing_token"
	errUnknownToken     = "unknown_token"
	errInvalidToken     = "invalid_token"
	errMissingParameter = "missing_parameter"
	errReadingFile      = "error_reading_file"
	errParsingApiBody   = "error_parsing_api_body"
)

func parseModelError(err models.IModelError) *failedApiResult {
	if err == nil {
		return nil
	}

	sc := 400
	code := ""
	switch err.ErrCode() {
	case models.ErrUnknownDBError:
		sc = 500
		code = errSystemError

	case models.ErrSystemError:
		sc = 500
		code = errSystemError

	default:
		code = string(err.ErrCode())
	}

	return newFailedApiResult(sc, code, err)
}
