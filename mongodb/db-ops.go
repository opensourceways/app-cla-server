package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func withContext1(f func(context.Context) dbmodels.IDBError) dbmodels.IDBError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}

func structToMap(info interface{}) (bson.M, dbmodels.IDBError) {
	body, err := util.BuildRequestBody(info, "")
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

func (this *client) pushArrayElem(ctx context.Context, collection, array string, filterOfDoc, value bson.M) dbmodels.IDBError {
	update := bson.M{"$push": bson.M{array: value}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
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
		return errNoDBRecord
	}
	return nil
}

func (this *client) replaceDoc(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, dbmodels.IDBError) {
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

func (this *client) deleteDoc(ctx context.Context, collection string, filterOfDoc bson.M) dbmodels.IDBError {
	col := this.collection(collection)
	if _, err := col.DeleteOne(ctx, filterOfDoc); err != nil {
		return newSystemError(err)
	}

	return nil
}

func (this *client) pullArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) dbmodels.IDBError {
	update := bson.M{"$pull": bson.M{array: filterOfArray}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

func (this *client) pushNestedArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M) dbmodels.IDBError {
	return this.updateArrayElemHelper(ctx, collection, array, filterOfDoc, filterOfArray, updateCmd, "$addToSet")
}

// r, _ := col.UpdateOne; r.ModifiedCount == 0 will happen in two case: 1. no matched array item; 2 update repeatedly with same update cmd.
func (this *client) updateArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M) dbmodels.IDBError {
	return this.updateArrayElemHelper(ctx, collection, array, filterOfDoc, filterOfArray, updateCmd, "$set")
}

func (this *client) updateArrayElemHelper(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M, op string) dbmodels.IDBError {
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
		bson.M{op: cmd},
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
		return errNoDBRecord
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
			return errNoDBRecord
		}
		return newSystemError(err)
	}
	return nil
}

func (this *client) moveArrayElem(ctx context.Context, collection, from, to string, filterOfDoc, filterOfArray, value bson.M) dbmodels.IDBError {
	col := this.collection(collection)

	r, err := col.UpdateOne(
		ctx, filterOfDoc,
		bson.M{
			"$pull": bson.M{from: filterOfArray},
			"$push": bson.M{to: value},
		},
	)
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

func (this *client) getDoc(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) dbmodels.IDBError {
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
			return errNoDBRecord
		}
		return newSystemError(err)
	}
	return nil
}

func (this *client) newDocIfNotExist(ctx context.Context, collection string, filterOfDoc, docInfo bson.M) (string, dbmodels.IDBError) {
	upsert := true

	col := this.collection(collection)
	r, err := col.UpdateOne(
		ctx, filterOfDoc, bson.M{"$setOnInsert": docInfo},
		&options.UpdateOptions{Upsert: &upsert},
	)
	if err != nil {
		return "", newSystemError(err)
	}

	if r.UpsertedID == nil {
		return "", newDBError(dbmodels.ErrRecordExists, fmt.Errorf("the doc exists"))
	}

	v, _ := toUID(r.UpsertedID)
	return v, nil
}

func (this *client) updateDoc(ctx context.Context, collection string, filterOfDoc, update bson.M) dbmodels.IDBError {
	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, bson.M{"$set": update})
	if err != nil {
		return newSystemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}
