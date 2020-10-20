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

const vcCollection = "verification_codes"

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
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(vcCollection)

		// delete the expired codes.
		filter := bson.M{"expiry": bson.M{"$lt": util.Now()}}
		col.DeleteMany(ctx, filter)

		// email + purpose can't be the index, for example: a corp signs two different communities
		// so, it should use insert doc instead
		_, err := c.insertDoc(ctx, vcCollection, body)
		return err
	}

	return c.doTransaction(f)
}

func (c *client) CheckVerificationCode(opt dbmodels.VerificationCode) error {
	var v struct {
		Expiry int64 `bson:"expiry"`
	}

	f := func(ctx context.Context) error {
		col := c.collection(vcCollection)

		filter := bson.M{
			"email":   opt.Email,
			"purpose": opt.Purpose,
			"code":    opt.Code,
		}
		opt := options.FindOneAndDeleteOptions{
			Projection: bson.M{"expiry": 1},
		}

		sr := col.FindOneAndDelete(ctx, filter, &opt)
		err := sr.Decode(&v)
		if err != nil && isErrNoDocuments(err) {
			return dbmodels.DBError{
				ErrCode: util.ErrWrongVerificationCode,
				Err:     fmt.Errorf("wrong verification code"),
			}
		}

		return err
	}

	if err := withContext(f); err != nil {
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
