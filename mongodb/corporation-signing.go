package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCorpSigning(orgRepo *dbmodels.OrgRepo) bson.M {
	return docFilterOfEnabledLink(orgRepo)
}

func elemFilterOfCorpSigning(email string) bson.M {
	return filterOfCorpID(email)
}

func (c *client) SignAsCorporation(orgRepo *dbmodels.OrgRepo, info *dbmodels.CorporationSigningInfo) error {
	signing := dCorpSigning{
		CLALanguage:     "",
		CorpID:          genCorpID(info.AdminEmail),
		CorporationName: info.CorporationName,
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		Date:            info.Date,
		SigningInfo:     info.Info,
	}
	body, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfCorpSigning(orgRepo)
	arrayFilterByElemMatch(fieldSignings, false, elemFilterOfCorpSigning(info.AdminEmail), docFilter)

	f := func(ctx context.Context) error {
		return c.pushArrayElem(ctx, c.corpSigningCollection, fieldSignings, docFilter, body)
	}

	return withContext(f)
}

func (c *client) ListCorporationSigning(orgRepo *dbmodels.OrgRepo, language string) ([]dbmodels.CorporationSigningSummary, error) {
	elemFilter := bson.M{}
	if language != "" {
		elemFilter[fieldCLALanguage] = language
	}

	v, err := c.listCorpSigning(orgRepo, elemFilter, projectOfCorpSigningBasicInfo())
	if err != nil || len(v) == 0 {
		return nil, err
	}

	r := make([]dbmodels.CorporationSigningSummary, 0, len(v))
	for _, item := range v {
		r = append(r, toModelOfCorporationSigning(&item))
	}

	return r, nil
}

func (c *client) GetCorporationSigningSummary(orgRepo *dbmodels.OrgRepo, email string) (dbmodels.CorporationSigningSummary, error) {
	v, err := c.listCorpSigning(orgRepo, elemFilterOfCorpSigning(email), projectOfCorpSigningBasicInfo())
	if err != nil || len(v) == 0 {
		return dbmodels.CorporationSigningSummary{}, err
	}

	return toModelOfCorporationSigning(&v[0]), nil
}

func (c *client) GetCorporationSigningDetail(orgRepo *dbmodels.OrgRepo, email string) (dbmodels.CorporationSigningDetail, error) {
	project := projectOfCorpSigningBasicInfo()
	project[memberNameOfSignings("info")] = 1

	v, err := c.listCorpSigning(orgRepo, elemFilterOfCorpSigning(email), project)
	if err != nil || len(v) == 0 {
		return dbmodels.CorporationSigningDetail{}, err
	}

	return dbmodels.CorporationSigningDetail{
		CorporationSigningSummary: toModelOfCorporationSigning(&v[0]),
		Info:                      v[0].SigningInfo,
	}, nil
}

func (c *client) listCorpSigning(orgRepo *dbmodels.OrgRepo, elemFilter, project bson.M) ([]dCorpSigning, error) {
	docFilter := docFilterOfCorpSigning(orgRepo)

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, c.corpSigningCollection, fieldSignings, docFilter, elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	}
	return v[0].Signings, nil
}

func toModelOfCorporationSigning(cs *dCorpSigning) dbmodels.CorporationSigningSummary {
	return dbmodels.CorporationSigningSummary{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail:      cs.AdminEmail,
			AdminName:       cs.AdminName,
			CorporationName: cs.CorporationName,
			Date:            cs.Date,
		},
		PDFUploaded: cs.PDFUploaded,
	}
}

func projectOfCorpSigningBasicInfo() bson.M {
	return bson.M{
		memberNameOfSignings("admin_email"):  1,
		memberNameOfSignings("admin_name"):   1,
		memberNameOfSignings("corp_name"):    1,
		memberNameOfSignings("date"):         1,
		memberNameOfSignings("pdf_uploaded"): 1,
	}
}
