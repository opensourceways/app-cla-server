package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
)

func NewUserService(
	us userservice.UserService,
	repo repository.CorpSigning,
) UserService {
	return &userService{
		us:   us,
		repo: repo,
	}
}

type UserService interface {
	ChangePassword(cmd *CmdToChangePassword) error
	Login(cmd *CmdToLogin) (dto UserLoginDTO, err error)
}

type userService struct {
	us   userservice.UserService
	repo repository.CorpSigning
}

func (s *userService) ChangePassword(cmd *CmdToChangePassword) error {
	return s.us.ChangePassword(cmd.Id, cmd.OldOne, cmd.NewOne)
}

func (s *userService) Login(cmd *CmdToLogin) (dto UserLoginDTO, err error) {
	var u domain.User

	if cmd.Account != nil {
		u, err = s.us.LoginByAccount(cmd.LinkId, cmd.Account, cmd.Password)
	} else {
		u, err = s.us.LoginByEmail(cmd.LinkId, cmd.Email, cmd.Password)
	}

	if err != nil {
		return
	}

	cs, err := s.repo.Find(u.CorpSigningId)
	if err != nil {
		return
	}

	if dto.Role = cs.GetRole(u.EmailAddr); dto.Role == "" {
		err = domain.NewDomainError(domain.ErrorCodeUserWrongAccountOrPassword)

		s.us.Remove([]string{u.Id})

		return
	}

	dto.Email = u.EmailAddr.EmailAddr()
	dto.CorpName = cs.CorpName().CorpName()
	dto.CorpSigningId = u.CorpSigningId
	dto.InitialPWChanged = u.PasswordChaged

	return
}
