package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfCorpSigning(email string) bson.M {
	return filterOfCorpID(email)
}

func (c *client) SignCorpCLA(linkID string, info *dbmodels.CorpSigningCreateOpt) dbmodels.IDBError {
	signing := dCorpSigning{
		CLALanguage: info.CLALanguage,
		CorpID:      genCorpID(info.AdminEmail),
		CorpName:    info.CorporationName,
		AdminEmail:  info.AdminEmail,
		AdminName:   info.AdminName,
		Date:        info.Date,
		SigningInfo: info.Info,
	}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, false, elemFilterOfCorpSigning(info.AdminEmail), docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return c.pushArrayElem(ctx, c.corpSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func (this *client) ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, dbmodels.IDBError) {
	elemFilter := map[string]bson.M{
		fieldCorpManagers: {"role": dbmodels.RoleAdmin},
	}
	if language != "" {
		elemFilter[fieldSignings] = bson.M{fieldLang: language}
	}

	project := projectOfCorpSigning()
	project[memberNameOfCorpManager(fieldEmail)] = 1

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getMultiArrays(
			ctx, this.corpSigningCollection, docFilterOfSigning(linkID),
			elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	n := len(signings)
	if n == 0 {
		return nil, nil
	}

	admins := map[string]bool{}
	for _, item := range v[0].Managers {
		admins[item.Email] = true
	}

	r := make([]dbmodels.CorporationSigningSummary, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, dbmodels.CorporationSigningSummary{
			CorporationSigningBasicInfo: *toDBModelCorporationSigningBasicInfo(&signings[i]),
			AdminAdded:                  admins[signings[i].AdminEmail],
		})
	}

	return r, nil
}

func (this *client) IsCorpSigned(linkID, email string) (bool, dbmodels.IDBError) {
	signed := false
	f := func(ctx context.Context) dbmodels.IDBError {
		v, err := this.isArrayElemNotExists(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email),
		)
		if err != nil {
			return newSystemError(err)
		}
		signed = !v
		return nil
	}

	err := withContext1(f)
	return signed, err
}

func (this *client) GetCorpSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, dbmodels.IDBError) {
	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email),
			projectOfCorpSigning(), &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, nil
	}

	return toDBModelCorporationSigningBasicInfo(&(signings[0])), nil
}

func (this *client) GetCorpSigningDetail(linkID, email string) (*dbmodels.CLAInfo, *dbmodels.CorpSigningCreateOpt, dbmodels.IDBError) {
	pipeline := bson.A{
		bson.M{"$match": docFilterOfSigning(linkID)},
		bson.M{"$project": bson.M{
			fieldCLAInfos: 1,
			fieldSignings: arrayElemFilter(fieldSignings, filterOfCorpID(email)),
		}},
		bson.M{"$unwind": "$" + fieldSignings},
		bson.M{"$project": bson.M{
			fieldSignings: 1,
			fieldCLAInfos: arrayElemFilter(
				fieldCLAInfos,
				bson.M{fieldLang: fmt.Sprintf("$%s.%s", fieldSignings, fieldLang)},
			),
		}},
	}

	var v []struct {
		CLAInfos []DCLAInfo   `bson:"cla_infos"`
		Signings dCorpSigning `bson:"signings"`
	}
	f := func(ctx context.Context) error {
		col := this.collection(this.corpSigningCollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	if err := withContext(f); err != nil {
		return nil, nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, nil, errNoDBRecord
	}

	signing := &(v[0].Signings)
	if signing.CLALanguage == "" {
		return nil, nil, nil
	}

	clas := v[0].CLAInfos
	if len(clas) == 0 {
		return nil, nil, nil
	}

	info := &dbmodels.CorpSigningCreateOpt{
		CorporationSigningBasicInfo: *toDBModelCorporationSigningBasicInfo(signing),
		Info:                        signing.SigningInfo,
	}
	cla := &clas[0]
	return &dbmodels.CLAInfo{
		CLAHash: cla.CLAHash,
		Fields:  toModelOfCLAFields(cla.Fields),
	}, info, nil
}

func toDBModelCorporationSigningBasicInfo(cs *dCorpSigning) *dbmodels.CorporationSigningBasicInfo {
	return &dbmodels.CorporationSigningBasicInfo{
		CLALanguage:     cs.CLALanguage,
		AdminEmail:      cs.AdminEmail,
		AdminName:       cs.AdminName,
		CorporationName: cs.CorpName,
		Date:            cs.Date,
	}
}

func projectOfCorpSigning() bson.M {
	return bson.M{
		memberNameOfSignings(fieldEmail): 1,
		memberNameOfSignings(fieldName):  1,
		memberNameOfSignings(fieldCorp):  1,
		memberNameOfSignings(fieldDate):  1,
		memberNameOfSignings(fieldLang):  1,
	}
}
