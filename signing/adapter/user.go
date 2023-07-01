package adapter

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
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

func (adapter *userAdatper) Login(opt *models.CorporationManagerAuthentication) (
	models.CorpManagerLoginInfo, models.IModelError,
) {
	r := models.CorpManagerLoginInfo{}

	cmd, err := adapter.cmdToLogin(opt)
	if err != nil {
		return r, toModelError(err)
	}

	v, err := adapter.s.Login(&cmd)
	if err != nil {
		if code, ok := err.(errorCode); ok {
			if code.ErrorCode() == domain.ErrorCodeUserWrongAccountOrPassword {
				return r, models.NewModelError(
					models.ErrWrongIDOrPassword,
					errors.New("wrong account or password"),
				)
			}
		}

		return r, toModelError(err)
	}

	r.Role = v.Role
	r.Email = v.Email
	r.CorpName = v.CorpName
	r.SigningId = v.CorpSigningId
	r.InitialPWChanged = v.InitialPWChanged

	return r, nil
}

func (adapter *userAdatper) cmdToLogin(opt *models.CorporationManagerAuthentication) (
	cmd app.CmdToLogin, err error,
) {
	cmd.LinkId = opt.LinkID
	if cmd.Password, err = dp.NewPassword(opt.Password); err != nil {
		return
	}

	if cmd.Account, err = dp.NewAccount(opt.User); err != nil {
		cmd.Email, err = dp.NewEmailAddr(opt.User)
	}

	return
}
