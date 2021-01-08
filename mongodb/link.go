package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfLink(orgRepo *dbmodels.OrgRepo) bson.M {
	return bson.M{
		"platform":      orgRepo.Platform,
		"org_id":        orgRepo.OrgID,
		"repo_id":       orgRepo.RepoID,
		fieldLinkStatus: linkStatusReady,
	}
}

func (this *client) CreateLink(info *dbmodels.LinkCreateOption) (string, dbmodels.IDBError) {
	doc, err := toDocOfLink(info)
	if err != nil {
		return "", err
	}

	docFilter := docFilterOfLink(&info.OrgRepo)

	docID := ""
	f := func(ctx context.Context) dbmodels.IDBError {
		s, err := this.newDocIfNotExist1(ctx, this.linkCollection, docFilter, doc)
		if err != nil {
			return err
		}
		docID = s
		return nil
	}

	if err = withContext1(f); err != nil {
		return "", err
	}
	return docID, nil
}

func toDocOfLink(info *dbmodels.LinkCreateOption) (bson.M, dbmodels.IDBError) {
	opt := cLink{
		LinkID:     info.LinkID,
		Platform:   info.Platform,
		OrgID:      info.OrgID,
		RepoID:     info.RepoID,
		OrgAlias:   info.OrgAlias,
		Submitter:  info.Submitter,
		LinkStatus: linkStatusReady,
	}
	body, err := structToMap1(opt)
	if err != nil {
		return nil, err
	}

	orgEmail, err := toDocOfOrgEmail(&info.OrgEmail)
	if err != nil {
		return nil, err
	}
	body[fieldOrgEmail] = orgEmail

	convertCLAs := func(field string, v []dbmodels.CLACreateOption) dbmodels.IDBError {
		clas := make(bson.A, 0, len(v))
		for i := range v {
			m, err := toDocOfCLA(&v[i])
			if err != nil {
				return err
			}
			clas = append(clas, m)
		}

		body[field] = clas
		return nil
	}

	if len(info.IndividualCLAs) > 0 {
		if err := convertCLAs(fieldIndividualCLAs, info.IndividualCLAs); err != nil {
			return nil, err
		}
	}

	if len(info.CorpCLAs) > 0 {
		if err := convertCLAs(fieldCorpCLAs, info.CorpCLAs); err != nil {
			return nil, err
		}
	}

	return body, nil
}

func toDocOfCLA(cla *dbmodels.CLACreateOption) (bson.M, dbmodels.IDBError) {
	info := &dCLA{
		URL:  cla.URL,
		Text: cla.Text,
		DCLAInfo: DCLAInfo{
			Fields:           toDocOfCLAField(cla.Fields),
			Language:         cla.Language,
			CLAHash:          cla.CLAHash,
			OrgSignatureHash: cla.OrgSignatureHash,
		},
	}
	r, err := structToMap1(info)
	if err != nil {
		return nil, err
	}

	if cla.OrgSignature != nil {
		r[fieldOrgSignature] = *cla.OrgSignature
	}

	return r, nil
}

func toDocOfCLAField(fs []dbmodels.Field) []dField {
	if len(fs) == 0 {
		return nil
	}

	fields := make([]dField, 0, len(fs))
	for _, item := range fs {
		fields = append(fields, dField{
			ID:          item.ID,
			Title:       item.Title,
			Type:        item.Type,
			Description: item.Description,
			Required:    item.Required,
		})
	}
	return fields
}
