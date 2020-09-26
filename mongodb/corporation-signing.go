package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type corporationSigningDoc struct {
	corporationSigning

	PDFUploaded bool `bson:"pdf_uploaded" json:"pdf_uploaded"`
	AdminAdded  bool `bson:"admin_added" json:"admin_added"`
}

type corporationSigning struct {
	AdminEmail      string                   `bson:"admin_email" json:"admin_email" required:"true"`
	AdminName       string                   `bson:"admin_name" json:"admin_name" required:"true"`
	CorporationName string                   `bson:"corp_name" json:"corp_name" required:"true"`
	Enabled         bool                     `bson:"enabled" json:"enabled"`
	Date            string                   `bson:"date" json:"date" required:"true"`
	SigningInfo     dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`
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
		Enabled:         info.Enabled,
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

func (c *client) ListCorporationSigning(opt dbmodels.CorporationSigningListOption) (map[string][]dbmodels.CorporationSigningDetails, error) {
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
		return nil, fmt.Errorf("build options to list corporation signing failed, err:%v", err)
	}
	filter := bson.M(body)
	filterForCorpSigning(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldCorporations: 1,

				fieldCorpoManagers: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorpoManagers),
					"cond":  bson.M{"$eq": bson.A{"$$this.role", dbmodels.RoleAdmin}},
				}},
			}},
			bson.M{"$project": bson.M{
				corpSigningField("corp_name"):   1,
				corpSigningField("admin_email"): 1,
				corpSigningField("admin_name"):  1,
				corpSigningField("enabled"):     1,

				corpoManagerElemKey("email"): 1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of corporation signing: %v", err)
		}
		return nil
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	r := map[string][]dbmodels.CorporationSigningDetails{}

	for i := 0; i < len(v); i++ {
		cs := v[i].Corporations
		if cs == nil || len(cs) == 0 {
			continue
		}

		admins := map[string]bool{}
		for _, m := range v[i].CorporationManagers {
			admins[m.Email] = true
		}

		cs1 := make([]dbmodels.CorporationSigningDetails, 0, len(cs))
		for _, item := range cs {

			cs1 = append(cs1, dbmodels.CorporationSigningDetails{
				CorporationSigningInfo: toDBModelCorporationSigningInfo(item),
				AdministratorEnabled:   admins[item.AdminEmail],
			})
		}
		r[objectIDToUID(v[i].ID)] = cs1
	}

	return r, nil
}

func (c *client) UpdateCorporationSigning(claOrgID, adminEmail, corporationName string, opt dbmodels.CorporationSigningUpdateInfo) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to build options for updating corporation signing, err:%v", err)
	}
	if len(body) == 0 {
		return nil
	}

	info := bson.M{}
	for k, v := range body {
		info[fmt.Sprintf("%s.$[elem].%s", fieldCorporations, k)] = v
	}

	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForCorpSigning(filter)

		update := bson.M{"$set": info}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"elem.corp_name":   corporationName,
						"elem.admin_email": adminEmail,
					},
				},
			},
		}

		r, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to update corporation signing, doesn't match any record")
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to update corporation signing, impossible")
		}
		return nil
	}

	return withContext(f)
}

func (c *client) GetCorporationSigningDetail(platform, org, repo, email string) (dbmodels.CorporationSigningDetail, error) {
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
		return dbmodels.CorporationSigningDetail{}, err
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
					return toDBModelCorporationSigningDetail(objectIDToUID(item.ID), &item.Corporations[0]), nil
				}
			}
		}
		if bingo {
			return dbmodels.CorporationSigningDetail{}, err
		}
	}

	for _, item := range v {
		if len(item.Corporations) != 0 {
			return toDBModelCorporationSigningDetail(objectIDToUID(item.ID), &item.Corporations[0]), nil
		}
	}

	return dbmodels.CorporationSigningDetail{}, err
}

func toDBModelCorporationSigningInfo(info corporationSigningDoc) dbmodels.CorporationSigningInfo {
	return dbmodels.CorporationSigningInfo{
		CorporationName: info.CorporationName,
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		Enabled:         info.Enabled,
		Info:            info.SigningInfo,
	}
}

func toDBModelCorporationSigningDetail(claOrgID string, cs *corporationSigningDoc) dbmodels.CorporationSigningDetail {
	return dbmodels.CorporationSigningDetail{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail:      cs.AdminEmail,
			AdminName:       cs.AdminName,
			CorporationName: cs.CorporationName,
			Date:            cs.Date,
		},
		PDFUploaded: cs.PDFUploaded,
		AdminAdded:  cs.AdminAdded,
		CLAOrgID:    claOrgID,
	}
}
