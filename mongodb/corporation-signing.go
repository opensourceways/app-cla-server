package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/zengchen1024/cla-server/models"
)

type corporationSigning struct {
	SigningInfo signingInfo `bson:"signing_info"`
	Enabled     bool        `bson:"enabled"`
}

func corpoSigningKey(email string) string {
	return fmt.Sprintf("corporations.%s", emailSuffixToKey(email))
}

func (c *client) SignAsCorporation(info models.CorporationSigning) error {
	oid, err := toObjectID(info.CLAOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		k := corpoSigningKey(info.Email)
		info.Info["email"] = info.Email
		v := bson.M{k: bson.M{"signing_info": info.Info, "enabled": false}}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid, k: bson.M{"$exists": false}}, bson.M{"$set": v})
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as corporation, maybe it has signed")
		}
		return nil
	}

	return withContext(f)
}
