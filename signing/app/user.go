package app

import "github.com/opensourceways/app-cla-server/signing/domain/userservice"

func NewUserService(
	us userservice.UserService,
) UserService {
	return &userService{
		us: us,
	}
}

type UserService interface {
	ChangePassword(cmd *CmdToChangePassword) error
}

type userService struct {
	us userservice.UserService
}

func (s *userService) ChangePassword(cmd *CmdToChangePassword) error {
	return s.us.ChangePassword(cmd.Id, cmd.OldOne, cmd.NewOne)
}
