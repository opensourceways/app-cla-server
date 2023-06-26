package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewUserAdapter(s app.UserService) *userAdatper {
	return &userAdatper{s}
}

type userAdatper struct {
	s app.UserService
}

func (adapter *userAdatper) ChangePassword(
	index string, opt *models.CorporationManagerResetPassword,
) models.IModelError {
	cmd, err := adapter.cmdToChangePassword(index, opt)
	if err != nil {
		return toModelError(err)
	}

	if err = adapter.s.ChangePassword(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *userAdatper) cmdToChangePassword(
	index string, opt *models.CorporationManagerResetPassword,
) (cmd app.CmdToChangePassword, err error) {
	if cmd.OldOne, err = dp.NewPassword(opt.OldPassword); err != nil {
		return
	}

	if cmd.NewOne, err = dp.NewPassword(opt.NewPassword); err != nil {
		return
	}

	cmd.Id = index

	return
}
