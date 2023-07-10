package app

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CmdToVerifySMTPEmail struct {
	Code      string
	Platform  string
	EmailAddr dp.EmailAddr
}

func (cmd *CmdToVerifySMTPEmail) purpose() (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf(
			"%s, %s/%s %s",
			vcTypeSMTPEmail, cmd.Platform, cmd.Code, cmd.EmailAddr.EmailAddr(),
		),
	)
}

func (cmd *CmdToVerifySMTPEmail) emailCredential() domain.EmailCredential {
	return domain.EmailCredential{
		Addr:     cmd.EmailAddr,
		Token:    []byte(cmd.Code),
		Platform: cmd.Platform,
	}
}

type CmdToAuthorizeSMTPEmail struct {
	CmdToVerifySMTPEmail

	VerificationCode string
}

func (cmd *CmdToAuthorizeSMTPEmail) key() (domain.VerificationCodeKey, error) {
	p, err := cmd.purpose()
	if err != nil {
		return domain.VerificationCodeKey{}, err
	}

	return domain.NewVerificationCodeKey(cmd.VerificationCode, p), nil
}
