package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func toDocOfOrgEmail(info *dbmodels.OrgEmailCreateInfo) (bson.M, error) {
	opt := dOrgEmail{
		Email:    info.Email,
		Platform: info.Platform,
	}

	body, err := structToMap(opt)
	if err != nil {
		return nil, err
	}

	body[fieldToken] = info.Token
	return body, nil
}

func toDocOfOrgCLA(info *dbmodels.OrgCLACreateOption) (bson.M, error) {
	opt := cOrgCLA{
		OrgIdentity: orgIdentity(&info.OrgRepo),
		DOrgRepo: DOrgRepo{
			Platform: info.Platform,
			OrgID:    info.OrgID,
			RepoID:   info.RepoID,
		},
		OrgAlias:   info.OrgAlias,
		OrgEmail:   info.OrgEmail,
		Submitter:  info.Submitter,
		LinkStatus: linkStatusUnabled,
	}
	body, err := structToMap(opt)
	if err != nil {
		return nil, err
	}

	convertCLAs := func(field string, v []dbmodels.CLA) error {
		clas := make(bson.A, 0, len(v))
		for _, item := range v {
			m, err := toDocOfCLA(&item)
			if err != nil {
				return err
			}
			clas = append(clas, m)
		}

		body[field] = clas
		return nil
	}

	if len(info.IndividualCLAs) > 0 {
		convertCLAs(fieldIndividualCLAs, info.IndividualCLAs)
	}

	if len(info.CorpCLAs) > 0 {
		convertCLAs(fieldCorpCLAs, info.CorpCLAs)
	}

	return body, nil
}

func docFilterOfLink(orgRepo *dbmodels.OrgRepo) bson.M {
	return bson.M{
		fieldOrgIdentity: orgIdentity(orgRepo),
		fieldLinkStatus:  bson.M{"$ne": linkStatusDeleted},
	}
}

func (this *client) CreateLink(info *dbmodels.OrgCLACreateOption) (string, error) {
	doc, err := toDocOfOrgCLA(info)
	if err != nil {
		return "", err
	}

	docFilter := docFilterOfLink(&info.OrgRepo)

	docID := ""
	f := func(ctx context.Context) error {
		s, err := this.newDocIfNotExist(ctx, this.orgCLACollection, docFilter, doc)
		if err != nil {
			return err
		}
		docID = s
		return nil
	}

	if err = withContext(f); err != nil {
		return "", err
	}
	return docID, nil
}

func (this *client) DeleteLink(docID string) error {
	oid, err := toObjectID(docID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		_, err := this.deleteDoc(ctx, this.orgCLACollection, docFilterByID(oid))
		return err
	}
	return withContext(f)
}

func (this *client) Unlink(platform, org, repo, applyTo string) error {
	f := func(ctx context.Context) error {
		return this.updateDoc(
			ctx, this.orgCLACollection,
			docFilterOfLink(platform, org, repo),
			bson.M{fieldLinkStatus: linkStatusDeleted},
		)
	}

	return withContext(f)
}
