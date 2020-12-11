package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCorpSigning(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func elemFilterOfCorpSigning(email string) bson.M {
	return filterOfCorpID(email)
}

func (c *client) SignAsCorporation(linkID string, info *dbmodels.CorporationSigningOption) error {
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

	docFilter := docFilterOfCorpSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, false, elemFilterOfCorpSigning(info.AdminEmail), docFilter)

	f := func(ctx context.Context) error {
		return c.pushArrayElem(ctx, c.corpSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext(f)
}

func (this *client) ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, error) {
	elemFilter := bson.M{}
	if language != "" {
		elemFilter[fieldCLALang] = language
	}

	project := projectOfCorpSigning()
	project[corpManagerField("email")] = 1

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getMultiArrays(
			ctx, this.corpSigningCollection, docFilterOfCorpSigning(linkID),
			map[string]bson.M{
				fieldCorpManagers: {"role": dbmodels.RoleAdmin},
				fieldSignings:     elemFilter,
			},
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) != 1 || v[0].Signings == nil {
		return nil, nil
	}

	admins := map[string]bool{}
	for _, item := range v[0].Managers {
		admins[item.Email] = true
	}

	signings := v[0].Signings
	n := len(signings)
	r := make([]dbmodels.CorporationSigningSummary, 0, n)

	for i := 0; i < n; i++ {
		r = append(r, toModelOfCorpSigningSummary(&signings[i], admins[signings[i].AdminEmail]))
	}

	return r, nil
}

func (this *client) GetCorpSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, error) {
	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfCorpSigning(linkID), filterOfCorpID(email),
			projectOfCorpSigning(), &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) != 1 || v[0].Signings == nil {
		return nil, nil
	}

	detail := toModelOfCorpSigningSummary(&(v[0].Signings[0]), false)
	return &detail.CorporationSigningBasicInfo, nil
}

func (this *client) GetCorpSigningDetail(linkID, email string) (*dbmodels.CorporationSigningOption, error) {
	project := bson.M{
		fieldLinkID:   1,
		fieldSignings: 1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfCorpSigning(linkID), filterOfCorpID(email), project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) != 1 || v[0].Signings == nil {
		return nil, nil
	}

	signing := &(v[0].Signings[0])
	detail := toModelOfCorpSigningSummary(signing, false)
	return &dbmodels.CorporationSigningOption{
		CorporationSigningBasicInfo: detail.CorporationSigningBasicInfo,
		Info:                        signing.SigningInfo,
	}, nil
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
