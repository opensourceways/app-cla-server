package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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

func toDocOfOrgCLA(info *dbmodels.LinkCreateOption) (bson.M, error) {
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

func (this *client) CreateLink(info *dbmodels.LinkCreateOption) (string, error) {
	doc, err := toDocOfOrgCLA(info)
	if err != nil {
		return "", err
	}

	docFilter := docFilterOfLink(&info.OrgRepo)

	docID := ""
	f := func(ctx context.Context) error {
		s, err := this.newDocIfNotExist(ctx, this.linkCollection, docFilter, doc)
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

func (this *client) Unlink(orgRepo *dbmodels.OrgRepo) error {
	status := bson.M{fieldLinkStatus: linkStatusDeleted}
	docFilter := docFilterOfLink(orgRepo)

	f := func(ctx mongo.SessionContext) error {
		this.updateDoc(ctx, this.linkCollection, docFilter, status)
		this.updateDoc(ctx, this.corpManagerCollection, docFilter, status)
		this.updateDoc(ctx, this.corpSigningCollection, docFilter, status)
		this.updateDoc(ctx, this.individualSigningCollection, docFilter, status)
		return nil
	}

	return this.doTransaction(f)
}

func (this *client) ListLinks(opt *dbmodels.LinkListOption) ([]dbmodels.LinkInfo, error) {
	filter := bson.M{
		"platform":      opt.Platform,
		"org_id":        bson.M{"$in": opt.Orgs},
		fieldLinkStatus: bson.M{"$ne": linkStatusDeleted},
	}

	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
	}

	var v []cOrgCLA
	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.linkCollection, filter, project, &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	r := make([]dbmodels.LinkInfo, 0, len(v))
	for _, item := range v {
		r = append(r, dbmodels.LinkInfo{
			OrgDetail: dbmodels.OrgDetail{
				OrgInfo:  toModelOfOrgInfo(&item),
				OrgEmail: item.OrgEmail,
			},
			Submitter: item.Submitter,
		})
	}

	return r, nil
}

func toModelOfOrgInfo(doc *cOrgCLA) dbmodels.OrgInfo {
	return dbmodels.OrgInfo{
		OrgRepo: dbmodels.OrgRepo{
			Platform: doc.Platform,
			OrgID:    doc.OrgID,
			RepoID:   doc.RepoID,
		},
		OrgAlias: doc.OrgAlias,
	}
}
