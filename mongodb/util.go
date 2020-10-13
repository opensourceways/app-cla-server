package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func structToMap(info interface{}) (map[string]interface{}, error) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     err,
		}
	}
	return body, nil
}

func addCorporationID(email string, body map[string]interface{}) {
	body[fieldCorporationID] = util.EmailSuffix(email)
}

func isHasNotSigned(err error) bool {
	e, ok := dbmodels.IsDBError(err)
	return ok && e.ErrCode == util.ErrHasNotSigned
}

func toObjectID(uid string) (primitive.ObjectID, error) {
	v, err := primitive.ObjectIDFromHex(uid)
	if err != nil {
		return v, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     fmt.Errorf("can't convert to object id"),
		}
	}
	return v, err
}

func isErrNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

func (c *client) pushArryItems(ctx context.Context, collection, array string, filterOfDoc bson.M, value interface{}) (*mongo.UpdateResult, error) {
	update := bson.M{"$push": bson.M{array: bson.M{"$each": value}}}

	col := c.collection(collection)
	return col.UpdateOne(ctx, filterOfDoc, update)
}

func (c *client) pullArryItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) (*mongo.UpdateResult, error) {
	update := bson.M{"$pull": bson.M{array: filterOfArray}}

	col := c.collection(collection)
	return col.UpdateOne(ctx, filterOfDoc, update)
}

func (c *client) updateArryItem(collection string, filterOfDoc, filterOfArray, updateCmd bson.M, ctx context.Context) (*mongo.UpdateResult, error) {
	update := bson.M{"$set": updateCmd}

	updateOpt := options.UpdateOptions{
		ArrayFilters: &options.ArrayFilters{
			Filters: bson.A{
				filterOfArray,
			},
		},
	}

	col := c.collection(collection)
	return col.UpdateOne(ctx, filterOfDoc, update, &updateOpt)
}

func (c *client) getArrayItem(collection string, filterOfDoc, filterOfArray, project bson.M, ctx context.Context, result interface{}) error {
	pipeline := bson.A{
		bson.M{"$match": filterOfDoc},
		bson.M{"$project": filterOfArray},
		bson.M{"$project": project},
	}

	col := c.collection(collection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}
