package app

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CmdToCreateCodeForSigning struct {
	LinkId    string
	EmailAddr dp.EmailAddr
}

func (cmd *CmdToCreateCodeForSigning) purpose() (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf("sign %s, %s", cmd.LinkId, cmd.EmailAddr.EmailAddr()),
	)
}

type CmdToValidateCodeForSigning struct {
	CmdToCreateCodeForSigning
	Code string
}

type CmdToCreateCodeForEmailDomain struct {
	CorpSigningId string
	EmailAddr     dp.EmailAddr
}

func (cmd *CmdToCreateCodeForEmailDomain) purpose() (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf(
			"add email domain: %s, %s",
			cmd.CorpSigningId, cmd.EmailAddr.EmailAddr(),
		),
	)
}

type CmdToValidateCodeForEmailDomain struct {
	CmdToCreateCodeForEmailDomain
	Code string
}
