package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	ats "github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"
)

func NewAccessTokenService(
	at ats.AccessTokenService,
) AccessTokenService {
	return &accessTokenService{
		at: at,
	}
}

type AccessTokenKeyDTO = domain.AccessTokenKey
type CmdToValidateAccessToken = domain.AccessTokenKey

type AccessTokenService interface {
	Remove(tokenId string)
	Add(payload []byte) (k domain.AccessTokenKey, err error)
	ValidateAndRefresh(old domain.AccessTokenKey) (domain.AccessTokenKey, []byte, error)
}

type accessTokenService struct {
	at ats.AccessTokenService
}

func (s *accessTokenService) Remove(tokenId string) {
	s.at.Remove(tokenId)
}

func (s *accessTokenService) Add(payload []byte) (k domain.AccessTokenKey, err error) {
	return s.at.Add(payload)
}

func (s *accessTokenService) ValidateAndRefresh(old domain.AccessTokenKey) (domain.AccessTokenKey, []byte, error) {
	return s.at.ValidateAndRefresh(old)
}
