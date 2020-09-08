package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func individualSigningKey(email string) string {
	return fmt.Sprintf("%s.%s", fieldIndividuals, strings.ReplaceAll(email, ".", "_"))
}

func (c *client) SignAsIndividual(claOrgID string, info dbmodels.IndividualSigningInfo) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		k := individualSigningKey(info.Email)
		v := bson.M{k: info.Info}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid, k: bson.M{"$exists": false}}, bson.M{"$set": v})
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as individual, maybe he/she has signed")
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as individual, impossible")

		}
		return nil
	}

	return withContext(f)
}
