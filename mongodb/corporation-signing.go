package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zengchen1024/cla-server/dbmodels"
)

const corporationsID = "corporations"

type corporationSigning struct {
	AdminEmail      string      `bson:"admin_email"`
	AdminName       string      `bson:"admin_name"`
	CorporationName string      `bson:"corporation_name"`
	Enabled         bool        `bson:"enabled"`
	SigningInfo     signingInfo `bson:"info"`
}

func corpoSigningKey(field string) string {
	return fmt.Sprintf("%s.%s", corporationsID, field)
}

func (c *client) SignAsCorporation(claOrgID string, info dbmodels.CorporationSigningInfo) error {
	claOrg, err := c.GetCLAOrg(claOrgID)
	if err != nil {
		return err
	}

	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return fmt.Errorf("Failed to build body for signing as corporation, err:%v", err)
	}

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": bson.M{
				"platform": claOrg.Platform,
				"org_id":   claOrg.OrgID,
				"repo_id":  claOrg.RepoID,
				"apply_to": claOrg.ApplyTo,
				"enabled":  true,
			}},
			bson.M{"$project": bson.M{
				"count": bson.M{"$cond": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$%s", corporationsID)},
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", corporationsID),
						"cond":  bson.M{"$eq": bson.A{"$$this.corporation_id", info.CorporationID}},
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
				return fmt.Errorf("Failed to add info when signing as corporation, it has signed")
			}
		}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{corporationsID: bson.M(body)}})
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as corporation, the cla bound to org is not exist")
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as corporation, impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) ListCorporationsOfOrg(opt dbmodels.CorporationSigningListOption) (map[string][]dbmodels.CorporationSigningInfo, error) {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return nil, fmt.Errorf("build options to list corporation signing failed, err:%v", err)
	}
	filter := bson.M(body)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				corpoSigningKey("corporation_name"): 1,
				corpoSigningKey("admin_email"):      1,
				corpoSigningKey("admin_name"):       1,
				corpoSigningKey("enabled"):          1,
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

	r := map[string][]dbmodels.CorporationSigningInfo{}

	for i := 0; i < len(v); i++ {
		cs := v[i].Corporations
		if cs == nil || len(cs) == 0 {
			continue
		}

		cs1 := make([]dbmodels.CorporationSigningInfo, 0, len(cs))
		for _, item := range cs {
			cs1 = append(cs1, toDBModelCorporationSigningInfo(item))
		}
		r[objectIDToUID(v[i].ID)] = cs1
	}

	return r, nil
}

func (c *client) UpdateCorporationOfOrg(claOrgID, adminEmail, corporationName string, opt dbmodels.CorporationSigningUpdateInfo) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to build options for updating corporation signing, err:%v", err)
	}
	if len(body) == 0 {
		return nil
	}

	info := bson.M{}
	for k, v := range body {
		info[fmt.Sprintf("%s.$.%s", corporationsID, k)] = v
	}

	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{
			"_id": oid,

			corpoSigningKey("corporation_name"): corporationName,
			corpoSigningKey("admin_email"):      adminEmail,
		}

		update := bson.M{"$set": info}

		r, err := col.UpdateOne(ctx, filter, update)
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

func toDBModelCorporationSigningInfo(info corporationSigning) dbmodels.CorporationSigningInfo {
	return dbmodels.CorporationSigningInfo{
		CorporationName: info.CorporationName,
		AdminEmail:      info.AdminEmail,
		AdminName:       info.AdminName,
		Enabled:         info.Enabled,
		Info:            info.SigningInfo,
	}
}
