package adapter

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type errorCode interface {
	ErrorCode() string
}

var errorCodeMap = map[string]models.ModelErrCode{
	domain.ErrorCodeCorpAdminExists:             models.ErrNoLinkOrManagerExists,
	domain.ErrorCodeCorpSigningReSigning:       models.ErrNoLinkOrResigned,
	domain.ErrorCodeCorpSigningNotFound:         models.ErrUnsigned,
	domain.ErrorCodeCorpSigningCanNotDelete:     models.ErrCorpManagerExists,
	domain.ErrorCodeCorpPDFNotFound:             models.ErrUnuploaed,
	domain.ErrorCodeUserSamePassword:            models.ErrSamePassword,
	domain.ErrorCodeUserInvalidPassword:         models.ErrInvalidPassword,
	domain.ErrorCodeUserUnmatchedPassword:       models.ErrNoLinkOrNoManager,
	domain.ErrorCodeUserInvalidAccount:          models.ErrInvalidManagerID,
	domain.ErrorCodeUserFrozen:                 models.ErrUserLoginFrozen,
	domain.ErrorCodeUserNotExists:              models.ErrUserNotExists,
	domain.ErrorCodeEmployeeManagerExists:       models.ErrCorpManagerExists,
	domain.ErrorCodeEmployeeManagerTooMany:      models.ErrManyEmployeeManagers,
	domain.ErrorCodeEmployeeManagerNotSameCorp:  models.ErrNotSameCorp,
	domain.ErrorCodeEmployeeNotSameCorp:         models.ErrNotSameCorp,
	domain.ErrorCodeEmployeeManagerAdminAsManager: models.ErrAdminAsManager,
	domain.ErrorCodeEmployeeSigningReSigning:    models.ErrNoLinkOrResigned,
	domain.ErrorCodeEmployeeSigningNoManager:    models.ErrNoCorpEmployeeManager,
	domain.ErrorCodeEmployeeSigningNotFound:     models.ErrNoLinkOrUnsigned,
	domain.ErrorCodeCorpEmailDomainNotMatch:     models.ErrUnmatchedEmailDomain,
	domain.ErrorCodeIndividualSigningReSigning:  models.ErrNoLinkOrResigned,
	domain.ErrorCodeIndividualSigningCorpExists: models.ErrGoToSignEmployeeCLA,
	domain.ErrorCodeCLAExists:                   models.ErrCLAExists,
	domain.ErrorCodeCLACanNotRemove:            models.ErrCLAIsUsed,
	domain.ErrorCodeLinkNotExists:              models.ErrNoLink,
	domain.ErrorCodeLinkExists:                 models.ErrLinkExists,
	domain.ErrorCodeLinkCanNotRemove:           models.ErrLinkIsUsed,
	domain.ErrorCodeGmailNoRefreshToken:        models.ErrNoRefreshToken,
	domain.ErrorCodeVerificationCodeBusy:        models.ErrTooManyRequest,
	domain.ErrorCodeNoPermission:               models.ErrNoPermission,
}

func toModelError(err error) models.IModelError {
	code, ok := err.(errorCode)
	if !ok {
		fmt.Println("toModelError: unexpected error type:", err)
		return models.NewModelError(models.ErrSystemError, err)
	}

	return models.NewModelError(codeMap(code.ErrorCode()), err)
}

func errBadRequestParameter(err error) models.IModelError {
	code, ok := err.(errorCode)
	if !ok {
		return models.NewModelError(models.ErrBadRequestParameter, err)
	}

	return models.NewModelError(codeMap(code.ErrorCode()), err)
}

func codeMap(code string) models.ModelErrCode {
	if v, ok := errorCodeMap[code]; ok {
		return v
	}
	return models.ErrBadRequestParameter
}