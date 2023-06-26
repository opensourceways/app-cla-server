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
	Add(linkId, csId string, managers []domain.Manager) (map[string]string, error)
	Remove(string, []domain.Manager)
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

func (s *userService) Add(linkId, csId string, managers []domain.Manager) (pws map[string]string, err error) {
	j := 0
	pw := ""

	for i := range managers {
		item := &managers[i]

		if pw, err = s.add(linkId, csId, item); err != nil {
			if commonRepo.IsErrorDuplicateCreating(err) {
				err = domain.NewDomainError(domain.ErrorCodeUserExists)
			}

			j = i
			break
		}

		pws[item.Id] = pw
	}

	if err != nil && j > 0 {
		s.Remove(linkId, managers[:j])
	}

	return
}

func (s *userService) Remove(linkId string, managers []domain.Manager) {
	for i := range managers {
		a, err := managers[i].Account()
		if err != nil {
			continue
		}

		if err := s.repo.Remove(linkId, a); err != nil {
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

func (s *userService) add(linkId, csId string, manager *domain.Manager) (p string, err error) {
	p, err = s.password.New()
	if err != nil {
		return
	}

	v, err := s.encrypt.Ecrypt(p)
	if err != nil {
		return
	}

	pw, err := dp.NewPassword(v)
	if err != nil {
		return
	}

	a, err := manager.Account()
	if err != nil {
		return
	}

	err = s.repo.Add(&domain.User{
		LinkId:        linkId,
		Account:       a,
		Password:      pw,
		EmailAddr:     manager.EmailAddr,
		CorpSigningId: csId,
	})

	return
}
