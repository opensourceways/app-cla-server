package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfCorpSigning(email string) bson.M {
	return filterOfCorpID(email)
}

func (c *client) SignCorpCLA(linkID string, info *dbmodels.CorpSigningCreateOpt) dbmodels.IDBError {
	signing := dCorpSigning{
		CLALanguage:     info.CLALanguage,
		CorpID:          genCorpID(info.AdminEmail),
		CorporationName: info.CorporationName,
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		Date:            info.Date,
		SigningInfo:     info.Info,
	}
	doc, err := structToMap1(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, false, elemFilterOfCorpSigning(info.AdminEmail), docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return c.pushArrayElem1(ctx, c.corpSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func (this *client) ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, dbmodels.IDBError) {
	elemFilter := map[string]bson.M{
		fieldCorpManagers: {"role": dbmodels.RoleAdmin},
	}
	if language != "" {
		elemFilter[fieldSignings] = bson.M{fieldCLALang: language}
	}

	project := projectOfCorpSigning()
	project[memberNameOfCorpManager("email")] = 1

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
		return nil, errNoDBRecord1
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
		r = append(r, toDBModelCorporationSigningDetail(&signings[i], admins[signings[i].AdminEmail]))
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
		return nil, errNoDBRecord1
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, nil
	}

	detail := toDBModelCorporationSigningDetail(&(signings[0]), false)
	return &detail.CorporationSigningBasicInfo, nil
}

func (this *client) GetCorpSigningDetail(linkID, email string) (*dbmodels.CorpSigningCreateOpt, dbmodels.IDBError) {
	project := bson.M{
		memberNameOfSignings("admin_email"): 1,
		memberNameOfSignings("admin_name"):  1,
		memberNameOfSignings("corp_name"):   1,
		memberNameOfSignings("date"):        1,
		memberNameOfSignings("info"):        1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord1
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, nil
	}

	detail := toDBModelCorporationSigningDetail(&(signings[0]), false)

	return &dbmodels.CorpSigningCreateOpt{
		CorporationSigningBasicInfo: detail.CorporationSigningBasicInfo,
		Info:                        signings[0].SigningInfo,
	}, nil
}

func toDBModelCorporationSigningDetail(cs *dCorpSigning, adminAdded bool) dbmodels.CorporationSigningSummary {
	return dbmodels.CorporationSigningSummary{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			CLALanguage:     cs.CLALanguage,
			AdminEmail:      cs.AdminEmail,
			AdminName:       cs.AdminName,
			CorporationName: cs.CorporationName,
			Date:            cs.Date,
		},
		AdminAdded: adminAdded,
	}
}

func projectOfCorpSigning() bson.M {
	return bson.M{
		memberNameOfSignings("admin_email"): 1,
		memberNameOfSignings("admin_name"):  1,
		memberNameOfSignings("corp_name"):   1,
		memberNameOfSignings("date"):        1,
		memberNameOfCorpManager("email"):    1,
	}
}
