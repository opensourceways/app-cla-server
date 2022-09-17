package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfLink(orgRepo *dbmodels.OrgRepo) bson.M {
	return bson.M{
		fieldPlatform:   orgRepo.Platform,
		fieldOrg:        orgRepo.OrgID,
		fieldRepo:       orgRepo.RepoID,
		fieldLinkStatus: linkStatusReady,
	}
}

func (this *client) GetLinkID(orgRepo *dbmodels.OrgRepo) (string, dbmodels.IDBError) {
	var v cLink
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.linkCollection, docFilterOfLink(orgRepo), bson.M{fieldLinkID: 1}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return "", err
	}

	return v.LinkID, nil
}

func (this *client) CreateLink(info *dbmodels.LinkCreateOption) (string, dbmodels.IDBError) {
	doc, err := toDocOfLink(info)
	if err != nil {
		return "", err
	}

	docFilter := docFilterOfLink(&info.OrgRepo)

	docID := ""
	f := func(ctx context.Context) dbmodels.IDBError {
		s, err := this.newDocIfNotExist(ctx, this.linkCollection, docFilter, doc)
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

func (this *client) UpdateLinkEmail(info *dbmodels.LinkCreateOption) dbmodels.IDBError {
	orgEmail, err := toDocOfOrgEmail(&info.OrgEmail)
	if err != nil {
		return err
	}
	fmt.Println("toddd")
	status := bson.M{fieldOrgEmail: orgEmail}
	docFilter := bson.M{fieldLinkID: info.LinkID}
	f := func(ctx context.Context) dbmodels.IDBError {
		err := this.updateDoc(ctx, this.linkCollection, docFilter, status)
		if err != nil {
			return err
		}
		return nil
	}
	return withContext1(f)
}

func (this *client) Unlink(linkID string) dbmodels.IDBError {
	status := bson.M{fieldLinkStatus: linkStatusDeleted}
	docFilter := bson.M{fieldLinkID: linkID}

	f := func(ctx context.Context) dbmodels.IDBError {
		err := this.updateDoc(ctx, this.linkCollection, docFilter, status)
		if err != nil {
			return err
		}

		this.updateDoc(ctx, this.corpSigningCollection, docFilter, status)
		this.updateDoc(ctx, this.individualSigningCollection, docFilter, status)
		return nil
	}

	return withContext1(f)
}

func (this *client) GetOrgOfLink(linkID string) (*dbmodels.OrgInfo, dbmodels.IDBError) {
	var v cLink
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.linkCollection,
			bson.M{
				fieldLinkID:     linkID,
				fieldLinkStatus: linkStatusReady,
			},
			bson.M{
				fieldIndividualCLAs: 0,
				fieldCorpCLAs:       0,
				fmt.Sprintf("%s.%s", fieldOrgEmail, fieldToken): 0,
			}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	r := toModelOfOrgInfo(&v)
	return &r, nil
}

func (this *client) ListLinks(opt *dbmodels.LinkListOption) ([]dbmodels.LinkInfo, dbmodels.IDBError) {
	filter := bson.M{
		fieldPlatform:   opt.Platform,
		fieldOrg:        bson.M{"$in": opt.Orgs},
		fieldLinkStatus: linkStatusReady,
	}

	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
		fmt.Sprintf("%s.%s", fieldOrgEmail, fieldToken): 0,
	}

	return this.getAllLinks(filter, project)
}

func (this *client) GetAllLinks() ([]dbmodels.LinkInfo, dbmodels.IDBError) {
	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
		fieldOrgEmail:       0,
	}
	return this.getAllLinks(bson.M{fieldLinkStatus: linkStatusReady}, project)
}

func (this *client) getAllLinks(filter, project bson.M) ([]dbmodels.LinkInfo, dbmodels.IDBError) {
	var v []cLink
	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.linkCollection, filter, project, &v)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
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

func toDocOfLink(info *dbmodels.LinkCreateOption) (bson.M, dbmodels.IDBError) {
	opt := cLink{
		LinkID:     info.LinkID,
		Platform:   info.Platform,
		Org:        info.OrgID,
		Repo:       info.RepoID,
		OrgAlias:   info.OrgAlias,
		Submitter:  info.Submitter,
		LinkStatus: linkStatusReady,
	}
	body, err := structToMap(opt)
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
			Fields:   toDocOfCLAField(cla.Fields),
			Language: cla.Language,
			CLAHash:  cla.CLAHash,
		},
	}
	r, err := structToMap(info)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func toDocOfCLAField(fs []dbmodels.Field) []dField {
	if len(fs) == 0 {
		return nil
	}

	fields := make([]dField, 0, len(fs))
	for i := range fs {
		item := &fs[i]
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

func toModelOfOrgInfo(doc *cLink) dbmodels.OrgInfo {
	return dbmodels.OrgInfo{
		OrgRepo: dbmodels.OrgRepo{
			Platform: doc.Platform,
			OrgID:    doc.Org,
			RepoID:   doc.Repo,
		},
		OrgAlias: doc.OrgAlias,
		OrgEmail: doc.OrgEmail.Email,
	}
}
