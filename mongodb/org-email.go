package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zengchen1024/cla-server/dbmodels"
)

const orgEmailCollection = "org_emails"

func (c *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to create org email info: build body err:%v", err)
	}

	var r *mongo.UpdateResult

	f := func(ctx context.Context) error {
		col := c.collection(orgEmailCollection)

		filter := bson.M{"email": opt.Email}

		upsert := true

		update := bson.M{"$setOnInsert": bson.M(body)}

		r, err = col.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return fmt.Errorf("Failed to create org email info: write db err:%v", err)
		}

		return nil
	}

	err = withContext(f)
	if err != nil {
		return err
	}

	if r.MatchedCount == 0 && r.UpsertedCount == 0 {
		return fmt.Errorf("Failed to create org email info: impossible")
	}

	return nil
}
