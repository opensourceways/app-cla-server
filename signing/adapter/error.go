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

func codeMap(code string) models.ModelErrCode {
	switch code {
	case domain.ErrorCodeCorpSigningReSigning, domain.ErrorCodeEmployeeSigningReSigning:
		return models.ErrNoLinkOrResigned

	// corp admin
	case domain.ErrorCodeCorpAdminExists:
		return models.ErrNoLinkOrManagerExists

	// corp signing
	case domain.ErrorCodeCorpSigningNotFound:
		return models.ErrUnsigned

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
	case domain.ErrorCodeEmployeeSigningNoManager:
		return models.ErrNoCorpEmployeeManager

	case domain.ErrorCodeEmployeeSigningNotFound:
		return models.ErrNoLinkOrUnsigned

	// corp email domain
	case domain.ErrorCodeCorpEmailDomainNotMatch:
		return models.ErrUnmatchedEmailDomain

	default:
		return models.ErrBadRequestParameter
	}
}
