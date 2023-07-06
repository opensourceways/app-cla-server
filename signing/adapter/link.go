package adapter

import (
	"errors"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewLinkAdapter(
	s app.LinkService,
	cla *claAdatper,
) *linkAdatper {
	return &linkAdatper{
		cla: cla,
		s:   s,
	}
}

type linkAdatper struct {
	cla *claAdatper

	s app.LinkService
}

func (adapter *linkAdatper) GetLink(linkId string) (
	org dbmodels.OrgInfo, merr models.IModelError,
) {
	v, err := adapter.s.Find(linkId)
	if err != nil {
		merr = toModelError(err)

		return
	}

	org.OrgID = v.Org.Org
	org.Platform = v.Org.Platform
	org.OrgAlias = v.Org.Alias
	org.OrgEmail = v.Email.Addr.EmailAddr()
	org.OrgEmailPlatform = v.Email.Platform

	return
}

// GetLinkCLA
func (adapter *linkAdatper) GetLinkCLA(linkId, claId string) (
	org dbmodels.OrgInfo, cla dbmodels.CLAInfo, merr models.IModelError,
) {
	v, err := adapter.s.FindLinkCLA(&domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})
	if err != nil {
		merr = toModelError(err)

		return
	}

	org.OrgID = v.Org.Org
	org.Platform = v.Org.Platform
	org.OrgAlias = v.Org.Alias
	org.OrgEmail = v.Email.Addr.EmailAddr()
	org.OrgEmailPlatform = v.Email.Platform

	cla.CLAId = v.CLA.Id
	cla.CLAFile = v.CLA.LocalFile
	cla.CLALang = v.CLA.Language
	cla.Fields = adapter.toFields(v.CLA.Fileds)

	return
}

// ListCLAs
func (adapter *linkAdatper) ListCLAs(linkId, applyTo string) ([]dbmodels.CLADetail, models.IModelError) {
	t, err := dp.NewCLAType(applyTo)
	if err != nil {
		return nil, toModelError(err)
	}

	v, err := adapter.s.FindCLAs(&app.CmdToFindCLAs{
		LinkId: linkId,
		Type:   t,
	})
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]dbmodels.CLADetail, len(v))
	for i := range v {
		item := &v[i]

		detail := &r[i]
		detail.CLAHash = item.Id
		detail.Language = item.Language
		detail.Fields = adapter.toFields(item.Fileds)
	}

	return r, nil
}

func (adapter *linkAdatper) toFields(fields []domain.Field) []dbmodels.Field {
	r := make([]dbmodels.Field, len(fields))

	for i := range fields {
		item := fields[i]
		r[i] = dbmodels.Field{
			ID:          item.Id,
			Type:        item.Type.CLAFieldType(),
			Title:       item.Title,
			Required:    item.Required,
			Description: item.Desc,
		}
	}

	return r
}

// List
func (adapter *linkAdatper) List(platform string, orgs []string) ([]dbmodels.LinkInfo, models.IModelError) {
	v, err := adapter.s.List(&app.CmdToListLink{
		Platform: platform,
		Orgs:     orgs,
	})
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]dbmodels.LinkInfo, len(v))
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
func (adapter *linkAdatper) Remove(linkId string) models.IModelError {
	if err := adapter.s.Remove(linkId); err != nil {
		return toModelError(err)
	}

	return nil
}

// Add
func (adapter *linkAdatper) Add(submitter string, opt *models.LinkCreateOption) models.IModelError {
	cmd, err := adapter.cmdToAddLink(submitter, opt)
	if err != nil {
		return toModelError(err)
	}

	if err := adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *linkAdatper) cmdToAddLink(submitter string, opt *models.LinkCreateOption) (
	cmd app.CmdToAddLink, err error,
) {
	if (opt.IndividualCLA == nil) && (opt.CorpCLA == nil) {
		err = errors.New("no cla instance")

		return
	}

	if opt.IndividualCLA != nil {
		v, err1 := adapter.cla.cmdToAddCLA(opt.IndividualCLA, dbmodels.ApplyToIndividual)
		if err1 != nil {
			err = err1

			return
		}

		cmd.CLAs = append(cmd.CLAs, v)
	}

	if opt.CorpCLA != nil {
		v, err1 := adapter.cla.cmdToAddCLA(opt.CorpCLA, dbmodels.ApplyToCorporation)
		if err1 != nil {
			err = err1

			return
		}

		cmd.CLAs = append(cmd.CLAs, v)
	}

	if cmd.Email, err = dp.NewEmailAddr(opt.OrgEmail); err != nil {
		return
	}

	cmd.Org.Org = opt.OrgID
	cmd.Org.Alias = opt.OrgAlias
	cmd.Org.Platform = opt.Platform
	cmd.Submitter = submitter

	return
}
