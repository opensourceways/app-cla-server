package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewEmployeeSigningAdapter(s app.EmployeeSigningService) *employeeSigningAdatper {
	return &employeeSigningAdatper{s}
}

type employeeSigningAdatper struct {
	s app.EmployeeSigningService
}

func (adapter *employeeSigningAdatper) Sign(opt *models.EmployeeSigning) models.IModelError {
	cmd, err := adapter.cmdToSignEmployeeCLA(opt)
	if err != nil {
		return toModelError(err)
	}

	if err = adapter.s.Sign(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *employeeSigningAdatper) cmdToSignEmployeeCLA(opt *models.EmployeeSigning) (
	cmd app.CmdToSignEmployeeCLA, err error,
) {
	// TODO missing cla id
	cmd.CLA.CLAId = opt.CLALanguage
	if cmd.CLA.Language, err = dp.NewLanguage(opt.CLALanguage); err != nil {
		return
	}

	if cmd.Rep.Name, err = dp.NewName(opt.Name); err != nil {
		return
	}

	if cmd.Rep.EmailAddr, err = dp.NewEmailAddr(opt.Email); err != nil {
		return
	}

	cmd.CorpSigningId = opt.CorpSigningId

	cmd.AllSingingInfo = opt.Info

	return
}
