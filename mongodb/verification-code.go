package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func (this *client) CreateVerificationCode(opt dbmodels.VerificationCode) dbmodels.IDBError {
	info := cVerificationCode{
		Email:   opt.Email,
		Code:    opt.Code,
		Purpose: opt.Purpose,
		Expiry:  opt.Expiry,
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		col := this.collection(this.vcCollection)
		col.DeleteOne(ctx, bson.M{
			fieldEmail:   info.Email,
			fieldPurpose: info.Purpose,
		})
		// delete the expired codes.
		filter := bson.M{fieldExpiry: bson.M{"$lt": util.Now()}}
		col.DeleteMany(ctx, filter)
		// email + purpose can't be the index, for example: a corp signs a community concurrently.
		// so, it should use insertDoc to record each verification codes.
		_, err := this.insertDoc(ctx, this.vcCollection, body)
		if err != nil {
			return newSystemError(err)
		}
		return nil
	}

	return withContext1(f)
}

func (this *client) GetVerificationCode(opt *dbmodels.VerificationCode) dbmodels.IDBError {
	var v struct {
		Expiry int64 `bson:"expiry"`
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		col := this.collection(this.vcCollection)

		sr := col.FindOneAndDelete(
			ctx,
			bson.M{
				fieldEmail:   opt.Email,
				fieldPurpose: opt.Purpose,
				fieldCode:    opt.Code,
			},
			&options.FindOneAndDeleteOptions{
				Projection: bson.M{fieldExpiry: 1},
			},
		)

		err := sr.Decode(&v)
		if err == nil {
			return nil
		}
		if isErrNoDocuments(err) {
			return errNoDBRecord
		}
		return newSystemError(err)
	}

	if err := withContext1(f); err != nil {
		return err
	}

	opt.Expiry = v.Expiry
	return nil
}
