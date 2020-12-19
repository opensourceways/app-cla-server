package mongodb

import (
	"context"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (this *client) getDoc1(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) *dbmodels.DBError {
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
		if err == mongo.ErrNoDocuments {
			return errNoDBRecord
		}
		return systemError(err)
	}
	return nil
}

func (this *client) pushArrayElem(ctx context.Context, collection, array string, filterOfDoc, value bson.M) *dbmodels.DBError {
	update := bson.M{"$push": bson.M{array: value}}

	col := this.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return systemError(err)
	}

	if r.MatchedCount == 0 {
		return errNoDBRecord
	}
	return nil
}

func (this *client) pullAndReturnArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M, result interface{}) *dbmodels.DBError {
	col := this.collection(collection)
	sr := col.FindOneAndUpdate(
		ctx, filterOfDoc,
		bson.M{"$pull": bson.M{array: filterOfArray}},
		&options.FindOneAndUpdateOptions{
			Projection: bson.M{array: bson.M{"$elemMatch": filterOfArray}},
		})

	if err := sr.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return errNoDBRecord
		}
		return systemError(err)
	}
	return nil
}

func (this *client) isArrayElemNotExists(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) (bool, *dbmodels.DBError) {
	query := bson.M{array: bson.M{"$elemMatch": filterOfArray}}
	for k, v := range filterOfDoc {
		query[k] = v
	}

	var v struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := this.getDoc1(ctx, collection, query, bson.M{"_id": 1}, &v)
	if err == nil {
		return true, nil
	}

	if err == errNoDBRecord {
		return false, nil
	}

	return false, err
}
