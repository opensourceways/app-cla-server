package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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

func (this *client) HasLink(orgRepo *dbmodels.OrgRepo) (bool, error) {
	var v cLink
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection, docFilterOfLink(orgRepo), bson.M{"_id": 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return false, err
	}

	return true, nil
}

func (this *client) CreateLink(info *dbmodels.LinkCreateOption) (string, error) {
	doc, err := toDocOfLink(info)
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

func (this *client) Unlink(linkID string) error {
	status := bson.M{fieldLinkStatus: linkStatusDeleted}
	docFilter := bson.M{fieldLinkID: linkID}

	f := func(ctx mongo.SessionContext) error {
		this.updateDoc(ctx, this.linkCollection, docFilter, status)
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

	var v []cLink
	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.linkCollection, filter, project, &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	n := len(v)
	if n == 0 {
		return nil, nil
	}

	r := make([]dbmodels.LinkInfo, 0, n)
	for i := range v {
		item := &v[i]
		r = append(r, dbmodels.LinkInfo{
			LinkID:    item.LinkID,
			OrgInfo:   toModelOfOrgInfo(item),
			Submitter: item.Submitter,
		})
	}

	return r, nil
}

func (this *client) GetOrgOfLink(linkID string) (*dbmodels.OrgInfo, error) {
	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
	}

	var v cLink
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection, bson.M{fieldLinkID: linkID}, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return &dbmodels.OrgInfo{
		OrgRepo: dbmodels.OrgRepo{
			Platform: v.Platform,
			OrgID:    v.OrgID,
			RepoID:   v.RepoID,
		},
		OrgAlias: v.OrgAlias,
		OrgEmail: v.OrgEmail,
	}, nil
}

func toModelOfOrgInfo(doc *cLink) dbmodels.OrgInfo {
	return dbmodels.OrgInfo{
		OrgRepo: dbmodels.OrgRepo{
			Platform: doc.Platform,
			OrgID:    doc.OrgID,
			RepoID:   doc.RepoID,
		},
		OrgAlias: doc.OrgAlias,
		OrgEmail: doc.OrgEmail,
	}
}

func toDocOfLink(info *dbmodels.LinkCreateOption) (bson.M, error) {
	opt := cLink{
		LinkID:     info.LinkID,
		Platform:   info.Platform,
		OrgID:      info.OrgID,
		RepoID:     info.RepoID,
		OrgAlias:   info.OrgAlias,
		OrgEmail:   info.OrgEmail,
		Submitter:  info.Submitter,
		LinkStatus: linkStatusUnready,
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
