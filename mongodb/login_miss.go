package mongodb

import (
	"context"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (this *client) GetLoginMiss(linkID, account string) (*dbmodels.LoginMiss, dbmodels.IDBError) {
	result := new(dbmodels.LoginMiss)
	f := func(ctx context.Context) dbmodels.IDBError {
		ops := options.FindOne()
		filter := genLoginMissFilter(linkID, account)
		err := this.collection(this.loginMiss).FindOne(ctx, filter, ops).Decode(result)
		if err == nil {
			return nil
		}

		if isErrNoDocuments(err) {
			return errNoDBRecord
		}
		return newSystemError(err)
	}
	if err := withContext1(f); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *client) CreateLoginMiss(lm dbmodels.LoginMiss) dbmodels.IDBError {
	bm, err := structToMap(lm)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		if _, err := this.collection(this.loginMiss).InsertOne(ctx, bm); err != nil {
			return newSystemError(err)
		}
		return nil
	}

	return withContext1(f)
}

func (this *client) UpdateLoginMiss(lm dbmodels.LoginMiss) dbmodels.IDBError {
	opt := options.Update().SetUpsert(true)
	filter := genLoginMissFilter(lm.LinkID, lm.Account)
	blm, IDBErr := structToMap(lm)
	if IDBErr != nil {
		return IDBErr
	}
	update := bson.M{"$set": blm}
	f := func(ctx context.Context) dbmodels.IDBError {
		result, err := this.collection(this.loginMiss).UpdateOne(ctx, filter, update, opt)
		if err != nil {
			return newSystemError(err)
		}
		if result.MatchedCount != 0 {
			beego.Info("matched and replaced an existing document")
		}
		if result.UpsertedCount != 0 {
			beego.Info(fmt.Printf("inserted a new document with ID %v\n", result.UpsertedID))
		}
		return nil
	}
	return withContext1(f)
}

func genLoginMissFilter(linkID, account string) bson.M {
	return bson.M{
		"link_id": linkID,
		"account": account,
	}
}
