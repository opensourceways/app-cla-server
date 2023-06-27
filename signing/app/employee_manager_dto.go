package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type CmdToAddEmployeeManager struct {
	CorpSigningId string
	Managers      []domain.Manager
}

type CmdToRemoveEmployeeManager struct {
	CorpSigningId string
	Managers      []string
}

type RemovedManagerDTO struct {
	Name  string
	Email string
}
