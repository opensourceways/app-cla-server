package adapter

import (
	"errors"

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
	org models.OrgInfo, merr models.IModelError,
) {
	v, err := adapter.s.Find(linkId)
	if err != nil {
		merr = toModelError(err)

		return
	}

	org.OrgAlias = v.Org.Alias
	org.OrgEmail = v.Email.Addr.EmailAddr()
	org.ProjectURL = v.Org.ProjectURL
	org.OrgEmailPlatform = v.Email.Platform

	return
}

// GetLinkCLA
func (adapter *linkAdatper) GetLinkCLA(linkId, claId string) (
	org models.OrgInfo, cla models.CLAInfo, merr models.IModelError,
) {
	v, err := adapter.s.FindLinkCLA(&domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})
	if err != nil {
		merr = toModelError(err)

		return
	}

	org.OrgAlias = v.Org.Alias
	org.OrgEmail = v.Email.Addr.EmailAddr()
	org.ProjectURL = v.Org.ProjectURL
	org.OrgEmailPlatform = v.Email.Platform

	cla.CLAId = v.CLA.Id
	cla.CLAFile = v.CLA.LocalFile
	cla.CLALang = v.CLA.Language
	cla.Fields = adapter.toFields(v.CLA.Fileds)

	return
}

// ListCLAs
func (adapter *linkAdatper) ListCLAs(linkId, applyTo string) ([]models.CLADetail, models.IModelError) {
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

	r := make([]models.CLADetail, len(v))
	for i := range v {
		item := &v[i]

		detail := &r[i]
		detail.CLAId = item.Id
		detail.Language = item.Language
		detail.Fields = adapter.toFields(item.Fileds)
	}

	return r, nil
}

func (adapter *linkAdatper) toFields(fields []domain.Field) []models.CLAField {
	r := make([]models.CLAField, len(fields))

	for i := range fields {
		item := fields[i]
		r[i] = models.CLAField{
			ID:          item.Id,
			Type:        item.Type,
			Title:       item.Title,
			Required:    item.Required,
			Description: item.Desc,
		}
	}

	return r
}

// List
func (adapter *linkAdatper) List(userId string) ([]models.LinkInfo, models.IModelError) {
	v, err := adapter.s.List(userId)
	if err != nil {
		return nil, toModelError(err)
	}

	r := make([]models.LinkInfo, len(v))
	for i := range v {
		item := &v[i]

		li := &r[i]

		li.LinkID = item.Id
		li.Submitter = item.Submitter

		li.OrgAlias = item.Org.Alias
		li.OrgEmail = item.Email.Addr.EmailAddr()
		li.ProjectURL = item.Org.ProjectURL
		li.OrgEmailPlatform = item.Email.Platform
	}

	return r, nil
}

// Remove
func (adapter *linkAdatper) Remove(userId, linkId string) models.IModelError {
	if err := adapter.s.Remove(userId, linkId); err != nil {
		return toModelError(err)
	}

	return nil
}

// Add
func (adapter *linkAdatper) Add(userId string, opt *models.LinkCreateOption) models.IModelError {
	cmd, err := adapter.cmdToAddLink(userId, opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err := adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *linkAdatper) cmdToAddLink(userId string, opt *models.LinkCreateOption) (
	cmd app.CmdToAddLink, err error,
) {
	if (opt.IndividualCLA == nil) && (opt.CorpCLA == nil) {
		err = errors.New("no cla instance")

		return
	}

	if opt.IndividualCLA != nil {
		opt.IndividualCLA.Type = models.ApplyToIndividual

		v, err1 := adapter.cla.cmdToAddCLA("", "", opt.IndividualCLA)
		if err1 != nil {
			err = err1

			return
		}

		cmd.CLAs = append(cmd.CLAs, v.CLAInfo)
	}

	if opt.CorpCLA != nil {
		opt.CorpCLA.Type = models.ApplyToCorporation

		v, err1 := adapter.cla.cmdToAddCLA("", "", opt.CorpCLA)
		if err1 != nil {
			err = err1

			return
		}

		cmd.CLAs = append(cmd.CLAs, v.CLAInfo)
	}

	if cmd.Email, err = dp.NewEmailAddr(opt.OrgEmail); err != nil {
		return
	}

	cmd.Org.Alias = opt.OrgAlias
	cmd.Org.ProjectURL = opt.ProjectURL

	cmd.Submitter = userId

	return
}
