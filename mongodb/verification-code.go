package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const verifCodeCollection = "verification_codes"

func (c *client) CreateVerificationCode(opt dbmodels.VerificationCode) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to build verification code body:%v", err)
	}

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(verifCodeCollection)

		// delete the old codes, including unused ones.
		filter := bson.M{"email": opt.Email, "purpose": opt.Purpose}
		col.DeleteMany(ctx, filter)

		upsert := true
		update := bson.M{"$setOnInsert": bson.M(body)}
		// add this filter in case of repetitive code
		filter["code"] = opt.Code

		r, err := col.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return fmt.Errorf("Failed to create verification code: write db err:%v", err)
		}

		if r.MatchedCount == 0 && r.UpsertedCount == 0 {
			return fmt.Errorf("Failed to create verification code: impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) CheckVerificationCode(opt dbmodels.VerificationCode) (bool, error) {
	valid := false

	f := func(ctx context.Context) error {
		col := c.collection(verifCodeCollection)

		filter := bson.M{
			"email":   opt.Email,
			"purpose": opt.Purpose,
			"code":    opt.Code,
		}
		opt := options.FindOneAndDeleteOptions{
			Projection: bson.M{"expiry": 1},
		}

		r := col.FindOneAndDelete(ctx, filter, &opt)

		var v struct {
			Expiry int64 `bson:"expiry"`
		}
		if err := r.Decode(&v); err != nil {
			if err.Error() == mongo.ErrNoDocuments.Error() {
				return nil
			}

			return fmt.Errorf("Failed to check verification code: %s", r.Err().Error())
		}

		valid = (v.Expiry >= time.Now().Unix())
		return nil
	}

	return valid, withContext(f)
}
