package adapter

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewEmployeeManagerAdapter(s app.EmployeeManagerService) *employeeManagerAdatper {
	return &employeeManagerAdatper{s}
}

type employeeManagerAdatper struct {
	s app.EmployeeManagerService
}

func (adapter *employeeManagerAdatper) Add(
	csId string, opt *models.EmployeeManagerCreateOption,
) (
	[]models.CorporationManagerCreateOption, models.IModelError,
) {
	cmd, me := adapter.cmdToAddEmployeeManager(csId, opt)
	if me != nil {
		return nil, me
	}

	dto, err := adapter.s.Add(&cmd)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.CorporationManagerCreateOption, len(dto))
	for i := range dto {
		r[i] = toCorporationManagerCreateOption(&dto[i])
	}

	return r, nil
}

func (adapter *employeeManagerAdatper) cmdToAddEmployeeManager(
	csId string, opt *models.EmployeeManagerCreateOption,
) (
	cmd app.CmdToAddEmployeeManager, me models.IModelError,
) {
	if len(opt.Managers) == 0 {
		me = models.NewModelError(
			models.ErrEmptyPayload, errors.New("no employee mangers"),
		)

		return
	}

	ids := map[string]bool{}
	emails := map[string]bool{}

	ms := make([]domain.Manager, len(opt.Managers))
	var err error
	for i := range opt.Managers {
		item := &opt.Managers[i]

		if ms[i], err = adapter.toManager(item); err != nil {
			me = toModelError(err)

			return
		}

		if ids[item.ID] {
			me = models.NewModelError(
				models.ErrDuplicateManagerID,
				errors.New("duplicate manager ID"),
			)

			return
		}
		ids[item.ID] = true

		if emails[item.Email] {
			me = models.NewModelError(
				models.ErrCorpManagerExists,
				errors.New("duplicate email"),
			)

			return
		}
		emails[item.Email] = true
	}

	cmd.Managers = ms
	cmd.CorpSigningId = csId

	return
}

func (adapter *employeeManagerAdatper) toManager(opt *models.EmployeeManager) (m domain.Manager, err error) {
	if m.Name, err = dp.NewName(opt.Name); err != nil {
		return
	}

	if m.EmailAddr, err = dp.NewEmailAddr(opt.Email); err != nil {
		return
	}

	m.Id = opt.ID

	_, err = m.Account()

	return
}

func toCorporationManagerCreateOption(dto *app.ManagerDTO) models.CorporationManagerCreateOption {
	return models.CorporationManagerCreateOption{
		ID:       dto.Account,
		Role:     dto.Role,
		Name:     dto.Name,
		Email:    dto.EmailAddr,
		Password: dto.Password,
	}
}

// Remove
func (adapter *employeeManagerAdatper) Remove(
	csId string, opt *models.EmployeeManagerDeleteOption,
) (
	[]models.CorporationManagerCreateOption, models.IModelError,
) {
	cmd, me := adapter.cmdToRemoveEmployeeManager(csId, opt)
	if me != nil {
		return nil, me
	}

	dto, err := adapter.s.Remove(&cmd)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.CorporationManagerCreateOption, len(dto))
	for i := range dto {
		item := &dto[i]

		r[i] = models.CorporationManagerCreateOption{
			Name:  item.Name,
			Email: item.Email,
		}
	}

	return r, nil
}

func (adapter *employeeManagerAdatper) cmdToRemoveEmployeeManager(
	csId string, opt *models.EmployeeManagerDeleteOption,
) (
	cmd app.CmdToRemoveEmployeeManager, me models.IModelError,
) {
	if len(opt.Managers) == 0 {
		me = models.NewModelError(
			models.ErrEmptyPayload, errors.New("no employee mangers"),
		)

		return
	}

	ids := make([]string, len(opt.Managers))
	for i := range opt.Managers {
		ids[i] = opt.Managers[i].ID
	}

	cmd.CorpSigningId = csId
	cmd.Managers = ids

	return
}

// List
func (adapter *employeeManagerAdatper) List(csId string) (
	[]models.CorporationManagerListResult, models.IModelError,
) {
	ms, err := adapter.s.List(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	v := make([]models.CorporationManagerListResult, len(ms))
	for i := range ms {
		v[i] = toCorporationManagerListResult(&ms[i])
	}

	return v, nil
}

func toCorporationManagerListResult(m *app.EmployeeManagerDTO) models.CorporationManagerListResult {
	return models.CorporationManagerListResult{
		ID:    m.ID,
		Name:  m.Name,
		Email: m.Email,
		Role:  models.RoleManager,
	}
}
