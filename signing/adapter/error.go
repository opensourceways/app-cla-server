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
	case domain.ErrorCodeCorpSigningReSigning:
		return models.ErrNoLinkOrResigned
	default:
		return models.ErrBadRequestParameter
	}
}
