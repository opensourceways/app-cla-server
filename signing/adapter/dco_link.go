package adapter

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewDCOLinkAdapter(
	s app.DCOLinkService,
	dco *dcoAdapter,
) *dcoLinkAdapter {
	return &dcoLinkAdapter{
		dco: dco,
		s:   s,
	}
}

type dcoLinkAdapter struct {
	dco *dcoAdapter

	s app.DCOLinkService
}

// ListDCOs
func (adapter *dcoLinkAdapter) ListDCOs(linkId string) ([]models.CLADetail, models.IModelError) {
	v, err := adapter.s.FindDCOs(linkId)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.CLADetail, len(v))
	for i := range v {
		item := &v[i]

		detail := &r[i]
		detail.CLAId = item.Id
		detail.Language = item.Language
		detail.Fields = toFields(item.Fileds)
	}

	return r, nil
}

// List
func (adapter *dcoLinkAdapter) List(platform string, orgs []string) ([]models.LinkInfo, models.IModelError) {
	v, err := adapter.s.List(&app.CmdToListLink{
		Platform: platform,
		Orgs:     orgs,
	})
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.LinkInfo, len(v))
	for i := range v {
		item := &v[i]

		li := &r[i]

		li.LinkID = item.Id
		li.Submitter = item.Submitter

		li.OrgID = item.Org.Org
		li.Platform = item.Org.Platform

		li.OrgAlias = item.Org.Alias
		li.OrgEmail = item.Email.Addr.EmailAddr()
		li.OrgEmailPlatform = item.Email.Platform
	}

	return r, nil
}

// Remove
func (adapter *dcoLinkAdapter) Remove(linkId string) models.IModelError {
	if err := adapter.s.Remove(linkId); err != nil {
		return toModelError(err)
	}

	return nil
}

// Add
func (adapter *dcoLinkAdapter) Add(submitter string, opt *models.DCOLinkCreateOption) models.IModelError {
	cmd, err := adapter.cmdToAddLink(submitter, opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err := adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *dcoLinkAdapter) cmdToAddLink(submitter string, opt *models.DCOLinkCreateOption) (
	cmd app.CmdToAddDCOLink, err error,
) {
	if opt.DCO == nil {
		err = errors.New("no dco instance")

		return
	}

	v, err := adapter.dco.cmdToAddDCO(opt.DCO)
	if err != nil {
		return
	}
	cmd.DCOs = append(cmd.DCOs, v)

	if cmd.Email, err = dp.NewEmailAddr(opt.OrgEmail); err != nil {
		return
	}

	cmd.Org.Org = opt.OrgID
	cmd.Org.Alias = opt.OrgAlias
	cmd.Org.Platform = opt.Platform
	cmd.Submitter = submitter

	return
}
