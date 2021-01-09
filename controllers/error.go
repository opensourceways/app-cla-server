package controllers

import "github.com/opensourceways/app-cla-server/models"

const (
	errSystemError          = "system_error"
	errMissingToken         = "missing_token"
	errUnknownToken         = "unknown_token"
	errInvalidToken         = "invalid_token"
	errMissingParameter     = "missing_parameter"
	errReadingFile          = "error_reading_file"
	errParsingApiBody       = "error_parsing_api_body"
	errResigned             = "resigned"
	errUnsigned             = string(models.ErrUnsigned)
	errNoLink               = string(models.ErrNoLink)
	errNoEmployeeManager    = "no_employee_manager"
	errWrongIDOrPassword    = "wrong_id_or_pw"
	errCorpManagerExists    = string(models.ErrCorpManagerExists)
	errNoRefreshToken       = "no_refresh_token"
	errUnknownEmailPlatform = "unknown_email_platform"
	errFileNotExists        = "file_not_exists"
	errLinkExists           = string(models.ErrLinkExists)
	errUnknownLink          = "unknown_link"
	errUnmatchedCLAType     = "unmatched_cla_type"
	errCLAExists            = string(models.ErrCLAExists)
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
