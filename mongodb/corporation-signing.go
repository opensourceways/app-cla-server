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

func (c *client) SignAsCorporation(linkID string, info *dbmodels.CorporationSigningOption) *dbmodels.DBError {
	signing := dCorpSigning{
		CLALanguage:     info.CLALanguage,
		CorpID:          genCorpID(info.AdminEmail),
		CorporationName: info.CorporationName,
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		Date:            info.Date,
		SigningInfo:     info.Info,
	}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, false, elemFilterOfCorpSigning(info.AdminEmail), docFilter)

	f := func(ctx context.Context) *dbmodels.DBError {
		return c.pushArrayElem(ctx, c.corpSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContextOfDB(f)
}

func (this *client) IsCorpSigned(linkID, email string) (bool, *dbmodels.DBError) {
	signed := false
	f := func(ctx context.Context) *dbmodels.DBError {
		v, err := this.isArrayElemNotExists(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email),
		)
		signed = v
		return err
	}

	err := withContextOfDB(f)
	return signed, err
}

func (this *client) ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, *dbmodels.DBError) {
	elemFilter := bson.M{}
	if language != "" {
		elemFilter[fieldCLALang] = language
	}

	project := projectOfCorpSigning()
	project[memberNameOfCorpManager("email")] = 1

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getMultiArrays(
			ctx, this.corpSigningCollection, docFilterOfSigning(linkID),
			map[string]bson.M{
				fieldCorpManagers: {"role": dbmodels.RoleAdmin},
				fieldSignings:     elemFilter,
			},
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	n := len(signings)
	if n == 0 {
		return nil, errNoChildDoc
	}

	admins := map[string]bool{}
	for _, item := range v[0].Managers {
		admins[item.Email] = true
	}

	r := make([]dbmodels.CorporationSigningSummary, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, toModelOfCorpSigningSummary(&signings[i], admins[signings[i].AdminEmail]))
	}

	return r, nil
}

func (this *client) GetCorpSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, *dbmodels.DBError) {
	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), filterOfCorpID(email),
			projectOfCorpSigning(), &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, errNoChildDoc
	}

	detail := toModelOfCorpSigningSummary(&(signings[0]), false)
	return &detail.CorporationSigningBasicInfo, nil
}

func (this *client) GetCorpSigningDetail(linkID, email string) ([]dbmodels.Field, *dbmodels.CorporationSigningOption, *dbmodels.DBError) {
	pipeline := bson.A{
		bson.M{"$match": docFilterOfSigning(linkID)},
		bson.M{"$project": bson.M{
			fieldSingingCLAInfo: 1,
			fieldSignings:       arrayElemFilter(fieldSignings, filterOfCorpID(email)),
		}},
		bson.M{"$unwind": "$" + fieldSignings},
		bson.M{"$project": bson.M{
			fieldSignings: 1,
			fieldSingingCLAInfo: arrayElemFilter(
				fieldSingingCLAInfo,
				bson.M{fieldCLALang: fmt.Sprintf("$%s.%s", fieldSignings, fieldCLALang)},
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
		return nil, nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, nil, errNoDBRecord
	}

	signing := &(v[0].Signings)
	if signing.CLALanguage == "" {
		return nil, nil, errNoChildDoc
	}

	clas := v[0].CLAInfos
	if len(clas) == 0 {
		return nil, nil, systemError(fmt.Errorf("impossible"))
	}

	detail := toModelOfCorpSigningSummary(signing, false)
	info := &dbmodels.CorporationSigningOption{
		CorporationSigningBasicInfo: detail.CorporationSigningBasicInfo,
		Info:                        signing.SigningInfo,
	}
	return toModelOfCLAFields(clas[0].Fields), info, nil
}

func toModelOfCorpSigningSummary(cs *dCorpSigning, adminAdded bool) dbmodels.CorporationSigningSummary {
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
		memberNameOfSignings(fieldCLALang):  1,
		memberNameOfSignings("admin_email"): 1,
		memberNameOfSignings("admin_name"):  1,
		memberNameOfSignings("corp_name"):   1,
		memberNameOfSignings("date"):        1,
	}
}
