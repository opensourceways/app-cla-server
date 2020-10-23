package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type corporationSigningDoc struct {
	AdminEmail      string                   `bson:"admin_email" json:"admin_email" required:"true"`
	AdminName       string                   `bson:"admin_name" json:"admin_name" required:"true"`
	CorporationName string                   `bson:"corp_name" json:"corp_name" required:"true"`
	Date            string                   `bson:"date" json:"date" required:"true"`
	SigningInfo     dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`

	PDFUploaded bool `bson:"pdf_uploaded" json:"pdf_uploaded"`
	AdminAdded  bool `bson:"admin_added" json:"admin_added"`

	PDF []byte `bson:"pdf" json:"pdf,omitempty"`
}

func filterForCorpSigning(filter bson.M) {
	filter["apply_to"] = dbmodels.ApplyToCorporation
	filter["enabled"] = true
}

func filterOfDocForCorpSigning(platform, org, repo string) bson.M {
	m, _ := filterOfOrgRepo(platform, org, repo)
	filterForCorpSigning(m)
	return m
}

func corpSigningField(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorporations, field)
}

func (c *client) SignAsCorporation(orgCLAID, platform, org, repo string, info dbmodels.CorporationSigningInfo) error {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return err
	}

	signing := corporationSigningDoc{
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		CorporationName: info.CorporationName,
		Date:            info.Date,
		SigningInfo:     info.Info,
	}
	body, err := structToMap(signing)
	if err != nil {
		return err
	}
	addCorporationID(info.AdminEmail, body)

	f := func(ctx mongo.SessionContext) error {
		notExist, err := c.isArrayElemNotExists(
			ctx, orgCLACollection, fieldCorporations,
			filterOfDocForCorpSigning(platform, org, repo),
			filterOfCorpID(info.AdminEmail),
		)
		if err != nil {
			return err
		}
		if !notExist {
			return dbmodels.DBError{
				ErrCode: util.ErrHasSigned,
				Err:     fmt.Errorf("this corp has already signed"),
			}
		}

		return c.pushArrayElem(ctx, orgCLACollection, fieldCorporations, filterOfDocID(oid), body)
	}

	return c.doTransaction(f)
}

func (c *client) ListCorporationSigning(opt dbmodels.CorporationSigningListOption) (map[string][]dbmodels.CorporationSigningDetail, error) {
	filterOfDoc, err := filterOfOrgRepo(opt.Platform, opt.OrgID, opt.RepoID)
	if err != nil {
		return nil, err
	}
	if opt.CLALanguage != "" {
		filterOfDoc["cla_language"] = opt.CLALanguage
	}
	filterForCorpSigning(filterOfDoc)
	filterOfDoc[fieldCorporations] = bson.M{"$type": "array"}

	var v []OrgCLA

	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, orgCLACollection, fieldCorporations,
			filterOfDoc, nil, projectOfCorpSigning(), &v)
	}

	if err = withContext(f); err != nil {
		return nil, err
	}

	r := map[string][]dbmodels.CorporationSigningDetail{}
	for _, doc := range v {
		cs := doc.Corporations
		if len(cs) == 0 {
			continue
		}

		cs1 := make([]dbmodels.CorporationSigningDetail, 0, len(cs))
		for _, item := range cs {
			cs1 = append(cs1, toDBModelCorporationSigningDetail(&item))
		}
		r[objectIDToUID(doc.ID)] = cs1
	}

	return r, nil
}

func (c *client) getCorporationSigningDetail(filterOfDoc bson.M, email string) (*OrgCLA, error) {
	var v []OrgCLA

	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, orgCLACollection, fieldCorporations,
			filterOfDoc, filterOfCorpID(email),
			projectOfCorpSigning(), &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return getSigningDoc(v, func(doc *OrgCLA) bool {
		return len(doc.Corporations) > 0
	})
}

func (c *client) GetCorporationSigningDetail(platform, org, repo, email string) (string, dbmodels.CorporationSigningDetail, error) {
	orgCLA, err := c.getCorporationSigningDetail(
		filterOfDocForCorpSigning(platform, org, repo), email,
	)
	if err != nil {
		return "", dbmodels.CorporationSigningDetail{}, err
	}

	return objectIDToUID(orgCLA.ID), toDBModelCorporationSigningDetail(&orgCLA.Corporations[0]), nil
}

func (c *client) CheckCorporationSigning(orgCLAID, email string) (dbmodels.CorporationSigningDetail, error) {
	var result dbmodels.CorporationSigningDetail

	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return result, err
	}

	orgCLA, err := c.getCorporationSigningDetail(filterOfDocID(oid), email)
	if err != nil {
		return result, err
	}

	return toDBModelCorporationSigningDetail(&orgCLA.Corporations[0]), nil
}

func toDBModelCorporationSigningDetail(cs *corporationSigningDoc) dbmodels.CorporationSigningDetail {
	return dbmodels.CorporationSigningDetail{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail:      cs.AdminEmail,
			AdminName:       cs.AdminName,
			CorporationName: cs.CorporationName,
			Date:            cs.Date,
		},
		PDFUploaded: cs.PDFUploaded,
		AdminAdded:  cs.AdminAdded,
	}
}

func projectOfCorpSigning() bson.M {
	return bson.M{
		corpSigningField("admin_email"):  1,
		corpSigningField("admin_name"):   1,
		corpSigningField("corp_name"):    1,
		corpSigningField("date"):         1,
		corpSigningField("pdf_uploaded"): 1,
		corpSigningField("admin_added"):  1,
	}
}
