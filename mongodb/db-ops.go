package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (this *client) pullAndReturnArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M, result interface{}) error {
	col := this.collection(collection)
	r := col.FindOneAndUpdate(
		ctx, filterOfDoc,
		bson.M{"$pull": bson.M{array: filterOfArray}},
		&options.FindOneAndUpdateOptions{
			Projection: bson.M{array: bson.M{"$elemMatch": filterOfArray}},
		})

	return r.Decode(result)
}
