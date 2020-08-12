package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/zengchen1024/cla-server/models"

	"go.mongodb.org/mongo-driver/bson"
)

func individualSigningKey(email string) string {
	return fmt.Sprintf("individuals.%s", strings.ReplaceAll(email, ".", "_"))
}

func (c *client) SignAsIndividual(info models.IndividualSigning) error {
	oid, err := toObjectID(info.CLAOrgID)
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
		return nil
	}

	return withContext(f)
}
