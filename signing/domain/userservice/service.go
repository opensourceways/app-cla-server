package userservice

import (
	"strings"

	"github.com/beego/beego/v2/core/logs"

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
	Add(linkId, csId string, managers []domain.Manager) (map[string]dp.Password, []string, error)
	Remove([]string)
	RemoveByAccount(linkId string, accounts []dp.Account)
	ChangePassword(index string, old, newOne dp.Password) error
	ResetPassword(linkId string, email dp.EmailAddr, newOne dp.Password) error
	Logout(userId string)
	LoginByAccount(linkId string, a dp.Account, p dp.Password) (domain.User, error)
	LoginByEmail(linkId string, e dp.EmailAddr, p dp.Password) (u domain.User, err error)
}

type userService struct {
	repo     repository.User
	encrypt  encryption.Encryption
	password userpassword.UserPassword
}

func (s *userService) Add(linkId, csId string, managers []domain.Manager) (map[string]dp.Password, []string, error) {
	ids := []string{}
	pws := map[string]dp.Password{}

	for i := range managers {
		item := &managers[i]

		pw, index, err := s.add(linkId, csId, item)
		if err != nil {
			if commonRepo.IsErrorDuplicateCreating(err) {
				err = domain.NewDomainError(domain.ErrorCodeUserExists)
			}

			if len(ids) > 0 {
				s.Remove(ids)
			}

			return nil, nil, err
		}

		pws[item.Id] = pw
		ids = append(ids, index)
	}

	return pws, ids, nil
}

func (s *userService) Remove(ids []string) {
	if err := s.repo.Remove(ids); err != nil {
		logs.Error(
			"remove user failed, user id: %s, err: %s",
			strings.Join(ids, ","), err.Error(),
		)
	}
}

func (s *userService) RemoveByAccount(linkId string, accounts []dp.Account) {
	if err := s.repo.RemoveByAccount(linkId, accounts); err != nil {
		v := make([]string, len(accounts))
		for i := range accounts {
			v[i] = accounts[i].Account()
		}

		logs.Error(
			"remove user failed, user: %s, err: %s",
			strings.Join(v, ","), err.Error(),
		)
	}
}

func (s *userService) ChangePassword(index string, old, newOne dp.Password) error {
	if err := s.checkPassword(newOne); err != nil {
		return err
	}

	if err := s.checkPassword(old); err != nil {
		return err
	}

	u, err := s.repo.Find(index)
	if err != nil {
		return err
	}

	err = u.ChangePassword(
		func(ciphertext []byte) bool {
			return s.isSamePassword(old, ciphertext)
		},

		func() ([]byte, error) {
			return s.encryptPassword(newOne)
		},
	)
	if err != nil {
		return err
	}

	return s.repo.SavePassword(&u)
}

func (s *userService) ResetPassword(linkId string, email dp.EmailAddr, newOne dp.Password) error {
	if err := s.checkPassword(newOne); err != nil {
		return err
	}

	u, err := s.repo.FindByEmail(linkId, email)
	if err != nil {
		return err
	}

	v, err := s.encryptPassword(newOne)
	if err != nil {
		return err
	}

	u.ResetPassword(v)

	return s.repo.SavePassword(&u)
}

func (s *userService) add(linkId, csId string, manager *domain.Manager) (p dp.Password, index string, err error) {
	p, err = s.password.New()
	if err != nil {
		return
	}

	v, err := s.encryptPassword(p)
	if err != nil {
		return
	}

	a, err := manager.Account()
	if err != nil {
		return
	}

	index, err = s.repo.Add(&domain.User{
		LinkId:        linkId,
		Account:       a,
		Password:      v,
		EmailAddr:     manager.EmailAddr,
		CorpSigningId: csId,
	})

	return
}

func (s *userService) checkPassword(p dp.Password) error {
	if !s.password.IsValid(p) {
		return domain.NewDomainError(domain.ErrorCodeUserInvalidPassword)
	}

	return nil
}

func (s *userService) Logout(userId string) {
	if u, err := s.repo.Find(userId); err == nil {
		u.Logout()

		if err := s.repo.SaveLoginInfo(&u); err != nil {
			logs.Error("save log out info, err:%s", err.Error())
		}
	}
}

func (s *userService) LoginByAccount(linkId string, a dp.Account, p dp.Password) (domain.User, error) {
	return s.login(
		func() (domain.User, error) {
			return s.repo.FindByAccount(linkId, a)
		},
		p,
	)
}

func (s *userService) LoginByEmail(linkId string, e dp.EmailAddr, p dp.Password) (u domain.User, err error) {
	return s.login(
		func() (domain.User, error) {
			return s.repo.FindByEmail(linkId, e)
		},
		p,
	)
}

func (s *userService) login(find func() (domain.User, error), p dp.Password) (u domain.User, err error) {
	loginErr := domain.NewDomainError(domain.ErrorCodeUserWrongAccountOrPassword)

	if !s.password.IsValid(p) {
		err = loginErr
		logs.Info("login 1")

		return
	}

	u, err = find()
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = loginErr
		}

		logs.Info("login 2")

		return
	}

	changed, err := u.Login(func(ciphertext []byte) bool {
		return s.isSamePassword(p, ciphertext)
	})

	if changed {
		logs.Info("login 3")

		if err1 := s.repo.SaveLoginInfo(&u); err1 != nil {
			logs.Error("save login info, err:%s", err1.Error())
		}
	}

	logs.Info("login 4")

	return
}

func (s *userService) isSamePassword(p dp.Password, ciphertext []byte) bool {
	return s.encrypt.IsSame(p.Password(), ciphertext)
}

func (s *userService) encryptPassword(p dp.Password) ([]byte, error) {
	return s.encrypt.Encrypt(p.Password())
}
