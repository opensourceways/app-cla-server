package app

import ats "github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"

func NewAccessTokenService(
	at ats.AccessTokenService,
) AccessTokenService {
	return &accessTokenService{
		at: at,
	}
}

type AccessTokenKeyDTO = ats.AccessTokenKey
type CmdToValidateAccessToken = ats.AccessTokenKey

type AccessTokenService interface {
	Add(payload []byte) (k ats.AccessTokenKey, err error)
	ValidateAndRefresh(old ats.AccessTokenKey) (ats.AccessTokenKey, []byte, error)
}

type accessTokenService struct {
	at ats.AccessTokenService
}

func (s *accessTokenService) Add(payload []byte) (k ats.AccessTokenKey, err error) {
	return s.at.Add(payload)
}

func (s *accessTokenService) ValidateAndRefresh(old ats.AccessTokenKey) (ats.AccessTokenKey, []byte, error) {
	return s.at.ValidateAndRefresh(old)
}
