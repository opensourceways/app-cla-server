package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignEmployeeCLA struct {
	CLA              domain.CLAInfo
	Rep              domain.Representative
	CorpSigningId    string
	AllSingingInfo   domain.AllSingingInfo
	VerificationCode string
}

func (cmd *CmdToSignEmployeeCLA) toEmployeeSigning() domain.EmployeeSigning {
	return domain.EmployeeSigning{
		CLA:     cmd.CLA,
		Rep:     cmd.Rep,
		Date:    util.Date(),
		AllInfo: cmd.AllSingingInfo,
	}
}

func (cmd *CmdToSignEmployeeCLA) toCmd() cmdToCreateCodeForEmployeeSigning {
	return cmdToCreateCodeForEmployeeSigning{
		Id:        cmd.CorpSigningId,
		EmailAddr: cmd.Rep.EmailAddr,
	}
}

// CmdToUpdateEmployeeSigning
type CmdToUpdateEmployeeSigning struct {
	CmdToRemoveEmployeeSigning

	Enabled bool
}

// CmdToRemoveEmployeeSigning
type CmdToRemoveEmployeeSigning struct {
	CorpSigningId     string
	EmployeeSigningId string
}

type IndividualSigningDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Date  string `json:"date"`
	Email string `json:"email"`
}

type EmployeeSigningDTO struct {
	IndividualSigningDTO

	Enabled bool `json:"enabled"`
}
