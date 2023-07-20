package loginservice

import (
	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/encryption"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userpassword"
)

var (
	errLogin  = domain.NewDomainError(domain.ErrorCodeUserWrongAccountOrPassword)
	errFrozen = domain.NewDomainError(domain.ErrorCodeUserFrozen)
)

func NewLoginService(
	user repository.User,
	repo repository.Login,
	encrypt encryption.Encryption,
	password userpassword.UserPassword,
) LoginService {
	return &loginService{
		user:     user,
		repo:     repo,
		encrypt:  encrypt,
		password: password,
	}
}

type LoginService interface {
	LoginByAccount(linkId string, a dp.Account, p dp.Password) (domain.User, domain.Login, error)
	LoginByEmail(linkId string, e dp.EmailAddr, p dp.Password) (domain.User, domain.Login, error)
}

type loginService struct {
	user     repository.User
	repo     repository.Login
	encrypt  encryption.Encryption
	password userpassword.UserPassword
}

func (s *loginService) LoginByAccount(linkId string, a dp.Account, p dp.Password) (domain.User, domain.Login, error) {
	return s.login(
		func() (domain.User, error) {
			return s.user.FindByAccount(linkId, a)
		},
		p, a.Account(),
	)
}

func (s *loginService) LoginByEmail(linkId string, e dp.EmailAddr, p dp.Password) (domain.User, domain.Login, error) {
	return s.login(
		func() (domain.User, error) {
			return s.user.FindByEmail(linkId, e)
		},
		p, e.EmailAddr(),
	)
}

func (s *loginService) login(find func() (domain.User, error), p dp.Password, lid string) (
	u domain.User, lv domain.Login, err error,
) {
	lv, err = s.repo.Find(lid)
	if err != nil {
		if !commonRepo.IsErrorResourceNotFound(err) {
			return
		}

		lv = domain.NewLogin(lid)
	}

	if lv.Frozen {
		err = errFrozen

		return
	}

	if !s.password.IsValid(p) {
		err = s.failToLogin(&lv)

		return
	}

	u, err = find()
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = s.failToLogin(&lv)
		}

		return
	}

	if s.isCorrectPassword(p, u.Password) {
		if lv.HasFailure() {
			if err1 := s.repo.Delete(lid); err1 != nil {
				logs.Error("delete login info failed, err:%s", err1.Error())
			}
		}

		return
	}

	err = s.failToLogin(&lv)

	return
}

func (s *loginService) failToLogin(l *domain.Login) error {
	err := errLogin
	if l.Fail() {
		err = errFrozen
	}

	if err1 := s.repo.Add(l); err1 != nil {
		logs.Error("save login info failed, err: %s", err1.Error())
	}

	return err
}

func (s *loginService) isCorrectPassword(p dp.Password, ciphertext []byte) bool {
	return s.encrypt.IsSame(p.Password(), ciphertext)
}
