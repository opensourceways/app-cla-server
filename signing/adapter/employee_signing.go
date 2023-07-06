package adapter

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
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

// Sign
func (adapter *employeeSigningAdatper) Sign(opt *models.EmployeeSigning) (
	[]dbmodels.CorporationManagerListResult, models.IModelError,
) {
	cmd, err := adapter.cmdToSignEmployeeCLA(opt)
	if err != nil {
		return nil, toModelError(err)
	}

	ms, err := adapter.s.Sign(&cmd)
	if err != nil {
		return nil, toModelError(err)
	}

	v := make([]dbmodels.CorporationManagerListResult, len(ms))
	for i := range ms {
		v[i] = toCorporationManagerListResult(&ms[i])
	}

	return v, nil
}

func (adapter *employeeSigningAdatper) cmdToSignEmployeeCLA(opt *models.EmployeeSigning) (
	cmd app.CmdToSignEmployeeCLA, err error,
) {
	cmd.CLA.CLAId = opt.CLAId
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

// Remove
func (adapter *employeeSigningAdatper) Remove(csId, esId string) (string, models.IModelError) {
	cmd := app.CmdToRemoveEmployeeSigning{}
	cmd.CorpSigningId = csId
	cmd.EmployeeSigningId = esId

	email, err := adapter.s.Remove(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return email, nil
}

// Update
func (adapter *employeeSigningAdatper) Update(csId, esId string, enabled bool) (string, models.IModelError) {
	cmd := app.CmdToUpdateEmployeeSigning{}
	cmd.CorpSigningId = csId
	cmd.EmployeeSigningId = esId
	cmd.Enabled = enabled

	email, err := adapter.s.Update(&cmd)
	if err != nil {
		return "", toModelError(err)
	}

	return email, nil
}

// List
func (adapter *employeeSigningAdatper) List(csId string) (
	[]dbmodels.IndividualSigningBasicInfo, models.IModelError,
) {
	v, err := adapter.s.List(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]dbmodels.IndividualSigningBasicInfo, len(v))
	for i := range v {
		item := &v[i]
		r[i] = dbmodels.IndividualSigningBasicInfo{
			ID:      item.ID,
			Name:    item.Name,
			Email:   item.Email,
			Date:    item.Date,
			Enabled: item.Enabled,
		}
	}

	return r, nil
}
