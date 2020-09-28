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

const verifCodeCollection = "verification_codes"

func (c *client) CreateVerificationCode(opt dbmodels.VerificationCode) error {
	info := struct {
		Email   string `json:"email" required:"true"`
		Code    string `json:"code" required:"true"`
		Purpose string `json:"purpose" required:"true"`
		Expiry  int64  `json:"expiry" required:"true"`
	}{
		Email:   opt.Email,
		Code:    opt.Code,
		Purpose: opt.Purpose,
		Expiry:  opt.Expiry,
	}

	body, err := structToMap(info)
	if err != nil {
		return fmt.Errorf("Failed to build verification code body:%v", err)
	}

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(verifCodeCollection)

		// delete the old codes, including unused ones.
		filter := bson.M{"email": opt.Email, "purpose": opt.Purpose}
		col.DeleteMany(ctx, filter)

		// add this filter in case of repetitive code
		filter["code"] = opt.Code

		upsert := true
		update := bson.M{"$setOnInsert": bson.M(body)}

		r, err := col.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return fmt.Errorf("write db err:%v", err)
		}

		if r.MatchedCount == 0 && r.UpsertedCount == 0 {
			return fmt.Errorf("impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) CheckVerificationCode(opt dbmodels.VerificationCode) error {
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
				return dbmodels.DBError{
					ErrCode: util.ErrWrongVerificationCode,
					Err:     fmt.Errorf("wrong verification code"),
				}
			}

			return err
		}

		if v.Expiry < util.Now() {
			return dbmodels.DBError{
				ErrCode: util.ErrVerificationCodeExpired,
				Err:     fmt.Errorf("verification code is expired"),
			}
		}
		return nil
	}

	return withContext(f)
}
