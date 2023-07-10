package app

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CmdToVerifyEmailDomain struct {
	CorpSigningId string
	EmailAddr     dp.EmailAddr
}

func (cmd *CmdToVerifyEmailDomain) purpose() (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf(
			"add email domain: %s, %s",
			cmd.CorpSigningId, cmd.EmailAddr.EmailAddr(),
		),
	)
}

type CmdToAddEmailDomain struct {
	CmdToVerifyEmailDomain

	VerificationCode string
}

func (cmd *CmdToAddEmailDomain) key() (domain.VerificationCodeKey, error) {
	p, err := cmd.purpose()
	if err != nil {
		return domain.VerificationCodeKey{}, err
	}

	return domain.NewVerificationCodeKey(cmd.VerificationCode, p), nil
}
