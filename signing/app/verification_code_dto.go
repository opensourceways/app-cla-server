package app

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type vcPurpose interface {
	purpose() (dp.Purpose, error)
}

// signing
type CmdToCreateCodeForSigning struct {
	LinkId    string
	EmailAddr dp.EmailAddr
}

func (cmd *CmdToCreateCodeForSigning) purpose() (dp.Purpose, error) {
	return cmd.newPurpose("signing")
}

func (cmd *CmdToCreateCodeForSigning) newPurpose(action string) (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf("%s, %s, %s", action, cmd.LinkId, cmd.EmailAddr.EmailAddr()),
	)
}

// email domain
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
