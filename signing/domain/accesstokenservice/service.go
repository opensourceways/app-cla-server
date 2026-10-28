package accesstokenservice

import (
	"encoding/base64"

	"github.com/opensourceways/obs-sdk/go/log"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/encryption"
	"github.com/opensourceways/app-cla-server/signing/domain/randombytes"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

var csrfTokenLen = 24

var invalidToken = domain.NewDomainError(domain.ErrorCodeAccessTokenInvalid)

type AccessTokenService interface {
	Remove(string)
	Add(payload []byte) (k domain.AccessTokenKey, err error)
	ValidateAndRefresh(old domain.AccessTokenKey) (newOne domain.AccessTokenKey, p []byte, err error)
}

func NewAccessTokenService(
	repo repository.AccessToken,
	encrypt encryption.Encryption,
	randomBytes randombytes.RandomBytes,
) AccessTokenService {
	return &accessTokenService{
		repo:        repo,
		encrypt:     encrypt,
		randomBytes: randomBytes,
	}
}

// accessTokenService
type accessTokenService struct {
	repo        repository.AccessToken
	encrypt     encryption.Encryption
	randomBytes randombytes.RandomBytes
}

func (s *accessTokenService) Remove(tokenId string) {
	if err := s.repo.Delete(tokenId); err != nil {
		log.Error("remove token failed", "error", err)
	}
}

func (s *accessTokenService) Add(payload []byte) (k domain.AccessTokenKey, err error) {
	bytes, err := s.randomBytes.New(csrfTokenLen)
	if err != nil {
		return
	}

	csrf, err := s.encryptToken(bytes)
	if err != nil {
		return
	}

	token := domain.NewAccessToken(payload, csrf)

	index, err := s.repo.Add(&token)
	if err != nil {
		return
	}

	k.Id = index
	k.CSRF = base64.StdEncoding.EncodeToString(bytes)

	return
}

func (s *accessTokenService) validate(old domain.AccessTokenKey) ([]byte, error) {
	csrf, err := base64.StdEncoding.DecodeString(old.CSRF)
	if err != nil || len(csrf) != csrfTokenLen {
		return nil, invalidToken
	}

	token, err := s.repo.Find(old.Id)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil, invalidToken
		}

		return nil, err
	}

	if !token.IsValid() {
		return nil, invalidToken
	}

	if !s.isSameToken(csrf, token.EncryptedCSRF) {
		return nil, invalidToken
	}

	return token.Payload, nil
}

func (s *accessTokenService) ValidateAndRefresh(old domain.AccessTokenKey) (
	newOne domain.AccessTokenKey, p []byte, err error,
) {
	if p, err = s.validate(old); err != nil {
		return
	}

	newOne, err1 := s.Add(p)
	if err1 == nil {
		if err1 := s.repo.Delete(old.Id); err1 != nil {
			log.Error("delete token failed", "token_id", old.Id, "error", err1)
		}
	} else {
		newOne = old
	}

	return
}

func (s *accessTokenService) isSameToken(plaintext, ciphertext []byte) bool {
	return s.encrypt.IsSame(plaintext, ciphertext)
}

func (s *accessTokenService) encryptToken(plaintext []byte) ([]byte, error) {
	return s.encrypt.Encrypt(plaintext)
}
