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

var (
	errNoDBRecord1 = dbError{code: dbmodels.ErrNoDBRecord, err: fmt.Errorf("no record")}
)

func withContext1(f func(context.Context) dbmodels.IDBError) dbmodels.IDBError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}

func structToMap1(info interface{}) (bson.M, dbmodels.IDBError) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, newDBError(dbmodels.ErrMarshalDataFaield, err)
	}
	return bson.M(body), nil
}

func arrayFilterByElemMatch(array string, exists bool, cond, filter bson.M) {
	match := bson.M{"$elemMatch": cond}
	if exists {
		filter[array] = match
	} else {
		filter[array] = bson.M{"$not": match}
	}
}

func (this *client) pushArrayElem1(ctx context.Context, collection, array string, filterOfDoc, value bson.M) dbmodels.IDBError {
	update := bson.M{"$push": bson.M{array: value}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord1
	}
	return nil
}

func (this *client) pushArrayElems(ctx context.Context, collection, array string, filterOfDoc bson.M, value bson.A) dbmodels.IDBError {
	update := bson.M{"$push": bson.M{array: bson.M{"$each": value}}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord1
	}
	return nil
}

func (this *client) replaceDoc1(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, dbmodels.IDBError) {
	upsert := true

	col := this.collection(collection)
	r, err := col.ReplaceOne(
		ctx, filterOfDoc, docInfo,
		&options.ReplaceOptions{Upsert: &upsert},
	)
	if err != nil {
		return "", newSystemError(err)
	}

	if r.UpsertedID == nil {
		return "", nil
	}

	v, _ := toUID(r.UpsertedID)
	return v, nil
}

func (this *client) pullArrayElem1(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) dbmodels.IDBError {
	update := bson.M{"$pull": bson.M{array: filterOfArray}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord1
	}
	return nil
}

// r, _ := col.UpdateOne; r.ModifiedCount == 0 will happen in two case: 1. no matched array item; 2 update repeatedly with same update cmd.
func (this *client) updateArrayElem1(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M) dbmodels.IDBError {
	cmd := bson.M{}
	for k, v := range updateCmd {
		cmd[fmt.Sprintf("%s.$[i].%s", array, k)] = v
	}

	arrayFilter := bson.M{}
	for k, v := range filterOfArray {
		arrayFilter["i."+k] = v
	}

	col := this.collection(collection)
	r, err := col.UpdateOne(
		ctx, filterOfDoc,
		bson.M{"$set": cmd},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					arrayFilter,
				},
			},
		},
	)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord1
	}
	return nil
}

func (this *client) pullAndReturnArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M, result interface{}) dbmodels.IDBError {
	col := this.collection(collection)
	sr := col.FindOneAndUpdate(
		ctx, filterOfDoc,
		bson.M{"$pull": bson.M{array: filterOfArray}},
		&options.FindOneAndUpdateOptions{
			Projection: bson.M{array: bson.M{"$elemMatch": filterOfArray}},
		})

	if err := sr.Decode(result); err != nil {
		if isErrNoDocuments(err) {
			return errNoDBRecord1
		}
		return newSystemError(err)
	}
	return nil
}

func (this *client) getDoc1(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) dbmodels.IDBError {
	col := this.collection(collection)

	var sr *mongo.SingleResult
	if len(project) > 0 {
		sr = col.FindOne(ctx, filterOfDoc, &options.FindOneOptions{
			Projection: project,
		})
	} else {
		sr = col.FindOne(ctx, filterOfDoc)
	}

	if err := sr.Decode(result); err != nil {
		if isErrNoDocuments(err) {
			return errNoDBRecord1
		}
		return newSystemError(err)
	}
	return nil
}
