package userservice

import (
	"github.com/sirupsen/logrus"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/encryption"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userpassword"
)

func NewUserService(
	repo repository.User,
	encrypt encryption.Encryption,
	password userpassword.UserPassword,
) UserService {
	return &userService{
		repo:     repo,
		encrypt:  encrypt,
		password: password,
	}
}

type UserService interface {
	Add(csId string, managers []domain.Manager) (err error)
	Remove([]domain.Manager)
	FindByAccount(dp.Account, dp.Password) (domain.User, error)
	FindByEmail(dp.EmailAddr, dp.Password) (domain.User, error)
	IsValidPassword(p dp.Password) bool
	ChangePassword(u *domain.User, p dp.Password) error
}

type userService struct {
	repo     repository.User
	encrypt  encryption.Encryption
	password userpassword.UserPassword
}

func (s *userService) Add(csId string, managers []domain.Manager) (err error) {
	j := 0
	for i := range managers {
		item := &managers[i]

		if err = s.add(csId, item); err != nil {
			if commonRepo.IsErrorDuplicateCreating(err) {
				err = domain.NewDomainError(domain.ErrorCodeUserExists)
			}

			j = i
			break
		}
	}

	if err != nil && j > 0 {
		s.Remove(managers[:j])
	}

	return
}

func (s *userService) Remove(managers []domain.Manager) {
	for i := range managers {
		a, err := managers[i].Account()
		if err != nil {
			continue
		}

		if err := s.repo.Remove(a); err != nil {
			logrus.Errorf(
				"remove user failed, user: %s, err: %s",
				a.Account(), err.Error(),
			)
		}
	}
}

func (s *userService) FindByAccount(a dp.Account, p dp.Password) (u domain.User, err error) {
	v, err := s.encrypt.Ecrypt(p.Password())
	if err != nil {
		return
	}

	return s.repo.FindByAccount(a, v)
}

func (s *userService) FindByEmail(e dp.EmailAddr, p dp.Password) (u domain.User, err error) {
	v, err := s.encrypt.Ecrypt(p.Password())
	if err != nil {
		return
	}

	return s.repo.FindByEmail(e, v)
}

func (s *userService) IsValidPassword(p dp.Password) bool {
	return s.password.IsValid(p.Password())
}

func (s *userService) ChangePassword(u *domain.User, p dp.Password) error {
	v, err := s.encrypt.Ecrypt(p.Password())
	if err != nil {
		return err
	}

	pw, err := dp.NewPassword(v)
	if err != nil {
		return err
	}

	u.ChangePassword(pw)

	return s.repo.Save(u)
}

func (s *userService) add(csId string, manager *domain.Manager) error {
	p, err := s.password.New()
	if err != nil {
		return err
	}

	v, err := s.encrypt.Ecrypt(p)
	if err != nil {
		return err
	}

	pw, err := dp.NewPassword(v)
	if err != nil {
		return err
	}

	a, err := manager.Account()
	if err != nil {
		return err
	}

	return s.repo.Add(&domain.User{
		EmailAddr:     manager.EmailAddr,
		Account:       a,
		Password:      pw,
		CorpSigningId: csId,
	})
}
