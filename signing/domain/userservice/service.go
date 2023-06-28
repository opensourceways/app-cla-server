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
	Add(linkId, csId string, managers []domain.Manager) (map[string]string, []string, error)
	Remove([]string)
	RemoveByAccount(linkId string, accounts []dp.Account)
	ChangePassword(index string, old, newOne dp.Password) error
}

type userService struct {
	repo     repository.User
	encrypt  encryption.Encryption
	password userpassword.UserPassword
}

func (s *userService) Add(linkId, csId string, managers []domain.Manager) (pws map[string]string, ids []string, err error) {
	pw := ""
	index := ""

	for i := range managers {
		item := &managers[i]

		if pw, index, err = s.add(linkId, csId, item); err != nil {
			if commonRepo.IsErrorDuplicateCreating(err) {
				err = domain.NewDomainError(domain.ErrorCodeUserExists)
			}

			break
		}

		pws[item.Id] = pw
		ids = append(ids, index)
	}

	if err != nil && len(ids) > 0 {
		s.Remove(ids)
	}

	return
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
	u, err := s.repo.Find(index)
	if err != nil {
		return err
	}

	old1, err := s.checkPassword(old)
	if err != nil {
		return err
	}

	newOne1, err := s.checkPassword(newOne)
	if err != nil {
		return err
	}

	if err := u.ChangePassword(old1, newOne1); err != nil {
		return err
	}

	return s.repo.SavePassword(&u)
}

func (s *userService) add(linkId, csId string, manager *domain.Manager) (p string, index string, err error) {
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

	index, err = s.repo.Add(&domain.User{
		LinkId:        linkId,
		Account:       a,
		Password:      pw,
		EmailAddr:     manager.EmailAddr,
		CorpSigningId: csId,
	})

	return
}

func (s *userService) checkPassword(p dp.Password) (dp.Password, error) {
	if !s.password.IsValid(p.Password()) {
		return nil, domain.NewDomainError(domain.ErrorCodeUserInvalidPassword)
	}

	v, err := s.encrypt.Ecrypt(p.Password())
	if err != nil {
		return nil, err
	}

	return dp.NewPassword(v)
}
