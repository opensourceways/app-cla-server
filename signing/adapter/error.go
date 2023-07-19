package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type errorCode interface {
	ErrorCode() string
}

func toModelError(err error) models.IModelError {
	code, ok := err.(errorCode)
	if !ok {
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
	switch code {
	// corp admin
	case domain.ErrorCodeCorpAdminExists:
		return models.ErrNoLinkOrManagerExists

	// corp signing
	case domain.ErrorCodeCorpSigningReSigning:
		return models.ErrNoLinkOrResigned

	case domain.ErrorCodeCorpSigningNotFound:
		return models.ErrUnsigned

	case domain.ErrorCodeCorpSigningCanNotDelete:
		return models.ErrCorpManagerExists

	// corp pdf
	case domain.ErrorCodeCorpPDFNotFound:
		return models.ErrUnuploaed

	// user
	case domain.ErrorCodeUserSamePassword:
		return models.ErrSamePassword

	case domain.ErrorCodeUserInvalidPassword:
		return models.ErrInvalidPassword

	case domain.ErrorCodeUserUnmatchedPassword:
		return models.ErrNoLinkOrNoManager

	case domain.ErrorCodeUserInvalidAccount:
		return models.ErrInvalidManagerID

	case domain.ErrorCodeUserFrozen:
		return models.ErrUserLoginFrozen

	// employee manager
	case domain.ErrorCodeEmployeeManagerExists:
		return models.ErrCorpManagerExists

	case domain.ErrorCodeEmployeeManagerTooMany:
		return models.ErrManyEmployeeManagers

	case domain.ErrorCodeEmployeeManagerNotSameCorp:
		return models.ErrNotSameCorp

	case domain.ErrorCodeEmployeeManagerAdminAsManager:
		return models.ErrAdminAsManager

	// employee signing
	case domain.ErrorCodeEmployeeSigningReSigning:
		return models.ErrNoLinkOrResigned

	case domain.ErrorCodeEmployeeSigningNoManager:
		return models.ErrNoCorpEmployeeManager

	case domain.ErrorCodeEmployeeSigningNotFound:
		return models.ErrNoLinkOrUnsigned

	// corp email domain
	case domain.ErrorCodeCorpEmailDomainNotMatch:
		return models.ErrUnmatchedEmailDomain

	// individual signing
	case domain.ErrorCodeIndividualSigningReSigning:
		return models.ErrNoLinkOrResigned

	case domain.ErrorCodeIndividualSigningCorpExists:
		return models.ErrGoToSignEmployeeCLA

	// cla
	case domain.ErrorCodeCLAExists:
		return models.ErrCLAExists

	// link
	case domain.ErrorCodeLinkNotExists:
		return models.ErrNoLink

	case domain.ErrorCodeLinkExists:
		return models.ErrLinkExists

	// gmail
	case domain.ErrorCodeGmailNoRefreshToken:
		return models.ErrNoRefreshToken

	// verification code
	case domain.ErrorCodeVerificationCodeBusy:
		return models.ErrTooManyRequest

	default:
		return models.ErrBadRequestParameter
	}
}
