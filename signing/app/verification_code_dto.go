package app

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type CmdToCreateCodeForSigning struct {
	LinkId    string
	EmailAddr dp.EmailAddr
}

func (cmd *CmdToCreateCodeForSigning) purpose() dp.Purpose {
	return dp.NewPurposeOfSigning(cmd.LinkId, cmd.EmailAddr)
}

type CmdToValidateCodeForSigning struct {
	CmdToCreateCodeForSigning
	Code string
}

type CmdToCreateCodeForEmailDomain struct {
	CorpSigningId string
	EmailAddr     dp.EmailAddr
}

func (cmd *CmdToCreateCodeForEmailDomain) purpose() dp.Purpose {
	return dp.NewPurposeOfSigning(cmd.CorpSigningId, cmd.EmailAddr)
}

type CmdToValidateCodeForEmailDomain struct {
	CmdToCreateCodeForEmailDomain
	Code string
}
