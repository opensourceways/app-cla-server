package mongodb

import (
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zengchen1024/cla-server/dbmodels"
)

type corporationSigning struct {
	SigningInfo signingInfo `bson:"info"`
	Enabled     bool        `bson:"enabled"`
}

func corpoSigningKey(email string) string {
	return fmt.Sprintf("corporations.%s", emailSuffixToKey(email))
}

func (c *client) SignAsCorporation(info dbmodels.CorporationSigningCreateOption) error {
	claOrg, err := c.GetCLAOrg(info.CLAOrgID)
	if err != nil {
		return err
	}

	oid, err := toObjectID(info.CLAOrgID)
	if err != nil {
		return err
	}

	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return fmt.Errorf("Failed to build body for signing as corporation, err:%v", err)
	}

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(claOrgCollection)

		pipline := bson.A{
			bson.M{"$match": bson.M{
				"corporations.admin_email": info.AdminEmail,
				"platform":                 claOrg.Platform,
				"org_id":                   claOrg.OrgID,
				"repo_id":                  claOrg.RepoID,
				"apply_to":                 claOrg.ApplyTo,
				"enabled":                  true,
			}},
			bson.M{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
		}

		cursor, err := col.Aggregate(ctx, pipline)
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

		if len(count) > 0 && count[0].Count != 0 {
			return fmt.Errorf("Failed to add info when signing as corporation, maybe it has signed")
		}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{"corporations": bson.M(body)}})
		if err != nil {
			return err
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as corporation, impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}
