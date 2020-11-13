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

func (this *client) CreateVerificationCode(opt dbmodels.VerificationCode) error {
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
		col := this.collection(this.vcCollection)

		// delete the expired codes.
		filter := bson.M{"expiry": bson.M{"$lt": util.Now()}}
		col.DeleteMany(ctx, filter)

		// email + purpose can't be the index, for example: a corp signs a community concurrently.
		// so, it should use insertDoc to record each verification codes.
		_, err := this.insertDoc(ctx, this.vcCollection, body)
		return err
	}

	return this.doTransaction(f)
}

func (this *client) GetVerificationCode(opt *dbmodels.VerificationCode) error {
	var v struct {
		Expiry int64 `bson:"expiry"`
	}

	f := func(ctx context.Context) error {
		col := this.collection(this.vcCollection)

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
				ErrCode: util.ErrNoDBRecord,
				Err:     fmt.Errorf("no db record"),
			}
		}

		return err
	}

	if err := withContext(f); err != nil {
		return err
	}

	opt.Expiry = v.Expiry
	return nil
}
