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

	case domain.ErrorCodeCorpAdminExists:
		return models.ErrNoLinkOrManagerExists

	case domain.ErrorCodeCorpSigningNotFound:
		return models.ErrUnsigned

	case domain.ErrorCodeCorpPDFNotFound:
		return models.ErrUnuploaed

	case domain.ErrorCodeUserSamePassword:
		return models.ErrSamePassword

	case domain.ErrorCodeUserInvalidPassword:
		return models.ErrInvalidPassword

	case domain.ErrorCodeUserUnmatchedPassword:
		return models.ErrNoLinkOrNoManager

	case domain.ErrorCodeUserInvalidAccount:
		return models.ErrInvalidManagerID

	case domain.ErrorCodeEmployeeManagerExists:
		return models.ErrCorpManagerExists

	case domain.ErrorCodeEmployeeManagerTooMany:
		return models.ErrManyEmployeeManagers

	case domain.ErrorCodeEmployeeManagerNotSameCorp:
		return models.ErrNotSameCorp

	case domain.ErrorCodeEmployeeManagerAdminAsManager:
		return models.ErrAdminAsManager

	default:
		return models.ErrBadRequestParameter
	}
}
