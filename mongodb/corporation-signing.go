package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type corporationSigningDoc struct {
	corporationSigning

	PDF []byte `bson:"pdf"`
}

type corporationSigning struct {
	AdminEmail      string                   `bson:"admin_email" json:"admin_email" required:"true"`
	AdminName       string                   `bson:"admin_name" json:"admin_name" required:"true"`
	CorporationName string                   `bson:"corp_name" json:"corp_name" required:"true"`
	Date            string                   `bson:"date" json:"date" required:"true"`
	SigningInfo     dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`

	PDFUploaded bool `bson:"pdf_uploaded" json:"pdf_uploaded"`
	AdminAdded  bool `bson:"admin_added" json:"admin_added"`
}

func filterForCorpSigning(filter bson.M) {
	filter["apply_to"] = dbmodels.ApplyToCorporation
	filter["enabled"] = true
	filter[fieldCorporations] = bson.M{"$type": "array"}
}

func corpSigningField(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorporations, field)
}

func (c *client) SignAsCorporation(claOrgID string, info dbmodels.CorporationSigningInfo) error {
	claOrg, err := c.GetBindingBetweenCLAAndOrg(claOrgID)
	if err != nil {
		return err
	}

	oid, _ := toObjectID(claOrgID)

	signing := corporationSigning{
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
		col := c.collection(claOrgCollection)

		filter := bson.M{
			"platform": claOrg.Platform,
			"org_id":   claOrg.OrgID,
			"repo_id":  claOrg.RepoID,
		}
		filterForCorpSigning(filter)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				"count": bson.M{"$cond": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$%s", fieldCorporations)},
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", fieldCorporations),
						"cond":  bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(info.AdminEmail)}},
					}}},
					0,
				}},
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		var count []struct {
			Count int `bson:"count"`
		}
		err = cursor.All(ctx, &count)
		if err != nil {
			return err
		}

		for _, item := range count {
			if item.Count != 0 {
				return fmt.Errorf("this corporation has signed")
			}
		}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{fieldCorporations: bson.M(body)}})
		if err != nil {
			return err
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) ListCorporationSigning(opt dbmodels.CorporationSigningListOption) (map[string][]dbmodels.CorporationSigningDetail, error) {
	info := struct {
		Platform    string `json:"platform" required:"true"`
		OrgID       string `json:"org_id" required:"true"`
		RepoID      string `json:"repo_id"`
		CLALanguage string `json:"cla_language,omitempty"`
	}{
		Platform:    opt.Platform,
		OrgID:       opt.OrgID,
		RepoID:      opt.RepoID,
		CLALanguage: opt.CLALanguage,
	}

	body, err := structToMap(info)
	if err != nil {
		return nil, err
	}
	filter := bson.M(body)
	filterForCorpSigning(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				corpSigningField("admin_email"):  1,
				corpSigningField("admin_name"):   1,
				corpSigningField("corp_name"):    1,
				corpSigningField("date"):         1,
				corpSigningField("pdf_uploaded"): 1,
				corpSigningField("admin_added"):  1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return err
		}
		return nil
	}

	err = withContext(f)
	if err != nil {
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

func (c *client) UploadCorporationSigningPDF(claOrgID, adminEmail string, pdf []byte) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForCorpSigning(filter)

		update := bson.M{"$set": bson.M{
			fmt.Sprintf("%s.$[elem].pdf", fieldCorporations):          pdf,
			fmt.Sprintf("%s.$[elem].pdf_uploaded", fieldCorporations): true,
		}}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"elem.corp_id": util.EmailSuffix(adminEmail),
					},
				},
			},
		}

		r, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		if r.ModifiedCount == 0 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the corp signing record"),
			}
		}
		return nil
	}

	return withContext(f)
}

func (c *client) DownloadCorporationSigningPDF(claOrgID, email string) ([]byte, error) {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return nil, err
	}

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForCorpSigning(filter)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldCorporations: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorporations),
					"cond":  bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
				}},
			}},
			bson.M{"$project": bson.M{
				corpSigningField("pdf"):          1,
				corpSigningField("pdf_uploaded"): 1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrInvalidParameter,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}

	cs := v[0].Corporations
	if len(cs) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrInvalidParameter,
			Err:     fmt.Errorf("can't find the corp signing in this record"),
		}
	}

	item := cs[0]
	if !item.PDFUploaded {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrPDFHasNotUploaded,
			Err:     fmt.Errorf("pdf has not yet been uploaded"),
		}
	}

	return item.PDF, nil
}

func (c *client) GetCorporationSigningDetail(platform, org, repo, email string) (string, dbmodels.CorporationSigningDetail, error) {
	filter := bson.M{
		"platform": platform,
		"org_id":   org,
	}
	if repo == "" {
		filter[fieldRepo] = ""
	} else {
		filter[fieldRepo] = bson.M{"$in": bson.A{"", repo}}

	}
	filterForCorpSigning(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldRepo: 1,
				fieldCorporations: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorporations),
					"cond":  bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
				}},
			}},
			bson.M{"$project": bson.M{
				fieldRepo:                        1,
				corpSigningField("admin_email"):  1,
				corpSigningField("admin_name"):   1,
				corpSigningField("corp_name"):    1,
				corpSigningField("date"):         1,
				corpSigningField("pdf_uploaded"): 1,
				corpSigningField("admin_added"):  1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	err := withContext(f)
	if err != nil {
		return "", dbmodels.CorporationSigningDetail{}, err
	}

	err = dbmodels.DBError{
		ErrCode: dbmodels.ErrHasNotSigned,
		Err:     fmt.Errorf("the corp:%s has not signed for this org/repo: %s/%s/%s ", util.EmailSuffix(email), platform, org, repo),
	}

	if repo != "" {
		bingo := false

		for _, item := range v {
			if item.RepoID == repo {
				if !bingo {
					bingo = true
				}
				if len(item.Corporations) != 0 {
					return objectIDToUID(item.ID), toDBModelCorporationSigningDetail(&item.Corporations[0]), nil
				}
			}
		}
		if bingo {
			return "", dbmodels.CorporationSigningDetail{}, err
		}
	}

	for _, item := range v {
		if len(item.Corporations) != 0 {
			return objectIDToUID(item.ID), toDBModelCorporationSigningDetail(&item.Corporations[0]), nil
		}
	}

	return "", dbmodels.CorporationSigningDetail{}, err
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
