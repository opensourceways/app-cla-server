package app

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CmdToVerifySMTPEmail struct {
	Code      []byte
	Platform  string
	EmailAddr dp.EmailAddr
}

func (cmd *CmdToVerifySMTPEmail) purpose() (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf(
			"%s %s %s",
			vcTypeSMTPEmail, cmd.Platform, cmd.EmailAddr.EmailAddr(),
		),
	)
}

func (cmd *CmdToVerifySMTPEmail) emailCredential() domain.EmailCredential {
	return domain.EmailCredential{
		Addr:     cmd.EmailAddr,
		Token:    cmd.Code,
		Platform: cmd.Platform,
	}
}

func (cmd *CmdToVerifySMTPEmail) clear() {
	for i := range cmd.Code {
		cmd.Code[i] = 0
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
