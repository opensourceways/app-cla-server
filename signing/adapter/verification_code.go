package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewVerificationCodeAdapter(s app.VerificationCodeService) *verificationCodeAdatper {
	return &verificationCodeAdatper{s}
}

type verificationCodeAdatper struct {
	s app.VerificationCodeService
}

// signing
func (adapter *verificationCodeAdatper) CreateForSigning(linkId string, email string) (
	string, models.IModelError,
) {
	cmd, err := adapter.toCmdToCreateCodeForSigning(linkId, email)
	if err != nil {
		return "", toModelError(err)
	}

	code, err := adapter.s.New(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return code, nil
}

func (adapter *verificationCodeAdatper) ValidateForSigning(linkId string, email, code string) models.IModelError {
	cmd, err := adapter.toCmdToCreateCodeForSigning(linkId, email)
	if err != nil {
		return toModelError(err)
	}

	if err := adapter.s.Validate(&cmd, code); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *verificationCodeAdatper) toCmdToCreateCodeForSigning(linkId string, email string) (
	cmd app.CmdToCreateCodeForSigning, err error,
) {
	cmd.LinkId = linkId
	cmd.EmailAddr, err = dp.NewEmailAddr(email)

	return
}
