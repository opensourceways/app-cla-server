package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewAccessTokenAdapter(
	s app.AccessTokenService,
) *accessTokenAdatper {
	return &accessTokenAdatper{s: s}
}

type accessTokenAdatper struct {
	s app.AccessTokenService
}

func (adapter *accessTokenAdatper) Remove(tokenId string) {
	adapter.s.Remove(tokenId)
}

func (adapter *accessTokenAdatper) Add(payload []byte) (models.AccessToken, models.IModelError) {
	v, err := adapter.s.Add(payload)
	if err != nil {
		return models.AccessToken{}, models.NewModelError(models.ErrSystemError, err)
	}

	return models.AccessToken{Id: v.Id, CSRF: v.CSRF}, nil
}

func (adapter *accessTokenAdatper) ValidateAndRefresh(old models.AccessToken) (
	newOne models.AccessToken, payload []byte, merr models.IModelError,
) {
	v, payload, err := adapter.s.ValidateAndRefresh(app.CmdToValidateAccessToken{
		Id:   old.Id,
		CSRF: old.CSRF,
	})

	if err != nil {
		code, ok := err.(errorCode)

		if ok && code.ErrorCode() == domain.ErrorCodeAccessTokenInvalid {
			merr = models.NewModelError(models.ErrInvalidToken, err)
		} else {
			merr = models.NewModelError(models.ErrSystemError, err)
		}

		return
	}

	newOne.Id = v.Id
	newOne.CSRF = v.CSRF

	return
}
