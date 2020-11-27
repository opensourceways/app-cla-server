package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func memberNameOfCorpCLA(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpCLAs, field)
}

func docFilterOfEnabledLink(orgRepo *dbmodels.OrgRepo) bson.M {
	return bson.M{
		fieldOrgIdentity: orgIdentity(orgRepo),
		fieldLinkStatus:  linkStatusEnabled,
	}
}

func fieldNameOfOrgEmailToken() string {
	return fmt.Sprintf("%s.%s", fieldOrgEmail, fieldToken)
}

func (this *client) GetLinkDetail(orgRepo *dbmodels.OrgRepo) (*dbmodels.OrgCLA, error) {
	var v cOrgCLA

	project := bson.M{
		fieldOrgIdentity:           0,
		fieldNameOfOrgEmailToken(): 0,
	}

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.orgCLACollection,
			docFilterOfLink(orgRepo),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return toModelOfOrgCLA(&v), nil
}

func (this *client) GetLinkByCLAType(orgRepo *dbmodels.OrgRepo, applyTo string) (*dbmodels.OrgCLA, error) {
	var v cOrgCLA

	project := bson.M{
		fieldOrgIdentity:           0,
		fieldNameOfOrgEmailToken(): 0,
	}

	if applyTo == dbmodels.ApplyToIndividual {
		project[fieldCorpCLAs] = 0

	} else {
		project[fieldIndividualCLAs] = 0
	}

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.orgCLACollection,
			docFilterOfLink(orgRepo),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return toModelOfOrgCLA(&v), nil
}

func (this *client) getLinkOfCorpCLA(orgRepo *dbmodels.OrgRepo, language, signatureMd5 string, result interface{}) error {
	fieldOfOrgSignature := memberNameOfCorpCLA(fieldOrgSignature)
	pipeline := bson.A{
		bson.M{"$match": docFilterOfEnabledLink(orgRepo)},
		bson.M{"$project": bson.M{
			fieldOrgAlias: 1,
			fieldOrgEmail: 1,
			fieldCorpCLAs: arrayElemFilter(fieldCorpCLAs, docFilterOfCLA(language)),
		}},
		bson.M{"$unwind": "$" + fieldCorpCLAs},
		bson.M{"$project": bson.M{
			fieldOrgAlias:                 1,
			fieldOrgEmail:                 1,
			memberNameOfCorpCLA("text"):   1,
			memberNameOfCorpCLA("fields"): 1,
			fieldOfOrgSignature: bson.M{"$cond": bson.M{
				"if":   bson.M{"$eq": bson.A{signatureMd5, "$" + memberNameOfCorpCLA(fieldOrgSignatureTag)}},
				"then": "$$REMOVE",
				"else": "$" + fieldOfOrgSignature,
			}},
		}},
	}

	f := func(ctx context.Context) error {
		col := this.collection(this.orgCLACollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, result)
	}

	return withContext(f)
}

func (this *client) GetLinkWhenSigningAsCorp(orgRepo *dbmodels.OrgRepo, language, signatureMd5 string) (*dbmodels.OrgCLA, error) {
	var v []struct {
		OrgAlias string `bson:"org_alias" json:"org_alias"`
		OrgEmail string `bson:"org_email" json:"-"`
		CorpCLA  dCLA   `bson:"corp_clas"`
	}
	err := this.getLinkOfCorpCLA(orgRepo, language, signatureMd5, &v)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrNoDBRecord,
			Err:     fmt.Errorf("no cla binding found"),
		}
	}

	doc := &v[0]
	r := &dbmodels.OrgCLA{
		OrgCLACreateOption: dbmodels.OrgCLACreateOption{
			OrgInfo: dbmodels.OrgInfo{
				OrgAlias: doc.OrgAlias,
			},
			OrgEmail: doc.OrgEmail,
			CorpCLAs: []dbmodels.CLA{*toModelOfCLA(&doc.CorpCLA)},
		},
	}
	return r, nil
}

func (this *client) GetLinkWhenSigningAsIndividual(orgRepo *dbmodels.OrgRepo, language string) (*dbmodels.OrgCLA, error) {
	var v []cOrgCLA
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.orgCLACollection, fieldIndividualCLAs,
			docFilterOfEnabledLink(orgRepo),
			docFilterOfCLA(language),
			bson.M{
				fieldOrgAlias: 1,
				fieldOrgEmail: 1,
				fmt.Sprintf("%s.%s", fieldIndividualCLAs, "fields"): 1,
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 || len(v[0].IndividualCLAs) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrNoDBRecord,
			Err:     fmt.Errorf("no cla binding found"),
		}
	}

	doc := &v[0]
	r := &dbmodels.OrgCLA{
		OrgCLACreateOption: dbmodels.OrgCLACreateOption{
			OrgInfo: dbmodels.OrgInfo{
				OrgAlias: doc.OrgAlias,
			},
			OrgEmail:       doc.OrgEmail,
			IndividualCLAs: []dbmodels.CLA{*toModelOfCLA(&doc.IndividualCLAs[0])},
		},
	}
	return r, nil
}

func (this *client) ListOrgs(opt *dbmodels.OrgListOption) ([]dbmodels.OrgInfo, error) {
	filter := bson.M{
		"platform":      opt.Platform,
		"org_id":        bson.M{"$in": opt.Orgs},
		fieldLinkStatus: bson.M{"$ne": linkStatusDeleted},
	}

	project := bson.M{
		fieldIndividualCLAs: 0,
		fieldCorpCLAs:       0,
		fieldOrgEmail:       0,
	}

	var v []cOrgCLA
	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.orgCLACollection, filter, project, &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	r := make([]dbmodels.OrgInfo, 0, len(v))
	for _, item := range v {
		r = append(r, toModelOfOrgInfo(&item))
	}

	return r, nil
}

func toModelOfOrgInfo(doc *cOrgCLA) dbmodels.OrgInfo {
	return dbmodels.OrgInfo{
		OrgRepo: dbmodels.OrgRepo{
			Platform: doc.Platform,
			OrgID:    doc.OrgID,
			RepoID:   toNormalRepo(doc.RepoID),
		},
		OrgAlias: doc.OrgAlias,
	}
}

func toModelOfOrgCLA(doc *cOrgCLA) *dbmodels.OrgCLA {
	r := &dbmodels.OrgCLA{
		OrgCLACreateOption: dbmodels.OrgCLACreateOption{
			OrgInfo:   toModelOfOrgInfo(doc),
			OrgEmail:  doc.OrgEmail,
			Submitter: doc.Submitter,
		},
	}

	convertCLAs := func(v []dCLA) []dbmodels.CLA {
		clas := make([]dbmodels.CLA, 0, len(v))
		for _, item := range v {
			clas = append(clas, *toModelOfCLA(&item))
		}

		return clas
	}

	if len(doc.IndividualCLAs) > 0 {
		r.IndividualCLAs = convertCLAs(doc.IndividualCLAs)
	}

	if len(doc.CorpCLAs) > 0 {
		r.CorpCLAs = convertCLAs(doc.CorpCLAs)
	}
	return r
}
