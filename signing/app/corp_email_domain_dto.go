package app

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type CmdToAddEmailDomain struct {
	EmailAddr     dp.EmailAddr
	CorpSigningId string
}
