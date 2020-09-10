package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const orgEmailCollection = "org_emails"

type OrgEmail struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Email    string             `bson:"email"`
	Platform string             `bson:"platform"`
	Token    []byte             `bson:"token"`
}

func (c *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to create org email info: build body err:%v", err)
	}
	body["token"] = opt.Token

	f := func(ctx context.Context) error {
		col := c.collection(orgEmailCollection)

		filter := bson.M{"email": opt.Email}
		upsert := true
		update := bson.M{"$setOnInsert": bson.M(body)}

		r, err := col.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return fmt.Errorf("Failed to create org email info: write db err:%v", err)
		}

		if r.MatchedCount == 0 && r.UpsertedCount == 0 {
			return fmt.Errorf("Failed to create org email info: impossible")
		}

		return nil
	}

	return withContext(f)
}

func (c *client) GetOrgEmailInfo(email string) (dbmodels.OrgEmailCreateInfo, error) {
	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.db.Collection(orgEmailCollection)
		opt := options.FindOneOptions{
			Projection: bson.M{"email": 0},
		}

		sr = col.FindOne(ctx, bson.M{"email": email}, &opt)
		return nil
	}

	r := dbmodels.OrgEmailCreateInfo{}
	err := withContext(f)
	if err != nil {
		return r, err
	}

	var v OrgEmail
	if err := sr.Decode(&v); err != nil {
		return r, fmt.Errorf("error decoding to bson struct: %s", err.Error())
	}

	return toDBModelOrgEmail(v), nil
}

func toDBModelOrgEmail(item OrgEmail) dbmodels.OrgEmailCreateInfo {
	return dbmodels.OrgEmailCreateInfo{
		Platform: item.Platform,
		Token:    item.Token,
	}
}
