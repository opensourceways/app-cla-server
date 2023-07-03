package accesstokenservice

import (
	"encoding/json"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/randomstr"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/symmetricencryption"
	"github.com/opensourceways/app-cla-server/util"
)

var invalidToken = domain.NewDomainError(domain.ErrorCodeAccessTokenInvalid)

type AccessTokenKey struct {
	Id   string
	CSRF string
}

type accessToken struct {
	CSRF    string
	Expiry  int64
	Payload []byte
}

func (at *accessToken) isValid(csrf string) bool {
	return at.CSRF == csrf && at.Expiry >= util.Now()
}

type AccessTokenService interface {
	Add(payload []byte) (k AccessTokenKey, err error)
	ValidateAndRefresh(old AccessTokenKey) (newOne AccessTokenKey, p []byte, err error)
}

func NewAccessTokenService(
	repo repository.AccessToken,
	expiry int64,
	encrypt symmetricencryption.Encryption,
	randomStr randomstr.RandomStr,
) AccessTokenService {
	return &accessTokenService{
		repo:      repo,
		expiry:    expiry,
		encrypt:   encrypt,
		randomStr: randomStr,
	}
}

// accessTokenService
type accessTokenService struct {
	repo      repository.AccessToken
	expiry    int64
	encrypt   symmetricencryption.Encryption
	randomStr randomstr.RandomStr
}

func (s *accessTokenService) Add(payload []byte) (k AccessTokenKey, err error) {
	code, err := s.randomStr.New()
	if err != nil {
		return
	}

	token := accessToken{
		CSRF:    code,
		Expiry:  s.expiry + util.Now(),
		Payload: payload,
	}

	v, err := json.Marshal(token)
	if err != nil {
		return
	}

	v, err = s.encrypt.Encrypt(v)
	if err != nil {
		return
	}

	index, err := s.repo.Add(v)
	if err != nil {
		return
	}

	k.Id = index
	k.CSRF = code

	return
}

func (s *accessTokenService) validate(old AccessTokenKey) ([]byte, error) {
	v, err := s.repo.Find(old.Id)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil, invalidToken
		}

		return nil, err
	}

	v, err = s.encrypt.Decrypt(v)
	if err != nil {
		return nil, err
	}

	token := accessToken{}

	if err = json.Unmarshal(v, &token); err != nil {
		return nil, err
	}

	if !token.isValid(old.CSRF) {
		return nil, invalidToken
	}

	return token.Payload, nil
}

func (s *accessTokenService) ValidateAndRefresh(old AccessTokenKey) (
	newOne AccessTokenKey, p []byte, err error,
) {
	if p, err = s.validate(old); err != nil {
		return
	}

	newOne, err1 := s.Add(p)
	if err1 == nil {
		if err1 := s.repo.Delete(old.Id); err1 != nil {
			logs.Error("delete token, id:%s, err:%s", old.Id, err1.Error())
		}
	} else {
		newOne = old
	}

	return
}
