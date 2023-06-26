package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type CmdToAddEmployeeManager struct {
	CorpSigningId string
	Managers      []domain.Manager
}
