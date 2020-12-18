package mongodb

import (
	"context"
	"fmt"

	"github.com/astaxie/beego"
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

func (this *client) HasLink(orgRepo *dbmodels.OrgRepo) (bool, *dbmodels.DBError) {
	var v cLink
	f := func(ctx context.Context) *dbmodels.DBError {
		return this.getDoc1(
			ctx, this.linkCollection, docFilterOfLink(orgRepo), bson.M{"_id": 1}, &v,
		)
	}

	if err := withContextOfDB(f); err != nil {
		return false, err
	}

	return true, nil
}

func (this *client) CreateLink(info *dbmodels.LinkCreateOption) (string, error) {
	beego.Info(" CreateLink")

	doc, err := toDocOfLink(info)
	if err != nil {
		return "", err
	}

	beego.Info(fmt.Sprintf("%#v", doc))
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
		fieldLinkStatus: linkStatusReady,
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

func (this *client) GetOrgOfLink(linkID string) (*dbmodels.OrgInfo, *dbmodels.DBError) {
	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
	}

	var v cLink
	f := func(ctx context.Context) *dbmodels.DBError {
		return this.getDoc1(
			ctx, this.linkCollection, bson.M{fieldLinkID: linkID}, project, &v,
		)
	}

	if err := withContextOfDB(f); err != nil {
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
	beego.Info(fmt.Sprintf("toDocOfLink: %#v", info))
	opt := cLink{
		LinkID:     info.LinkID,
		Platform:   info.Platform,
		OrgID:      info.OrgID,
		RepoID:     info.RepoID,
		OrgAlias:   info.OrgAlias,
		OrgEmail:   info.OrgEmail,
		Submitter:  info.Submitter,
		LinkStatus: linkStatusReady,
	}
	body, err := structToMap(opt)
	if err != nil {
		return nil, err
	}

	convertCLAs := func(field string, v []dbmodels.CLACreateOption) error {
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
