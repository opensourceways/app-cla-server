package mongodb

import (
	"context"
	"fmt"
	"strings"

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
	body[fieldCorporationID] = genCorpID(email)
}

func genCorpID(email string) string {
	return util.EmailSuffix(email)
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

func (c *client) updateArryItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M) (*mongo.UpdateResult, error) {
	cmd := bson.M{}
	for k, v := range updateCmd {
		cmd[fmt.Sprintf("%s.$[i].%s", array, k)] = v
	}
	update := bson.M{"$set": cmd}

	filterOfArray1 := bson.M{}
	for k, v := range filterOfArray {
		filterOfArray1["i."+k] = v
	}

	updateOpt := options.UpdateOptions{
		ArrayFilters: &options.ArrayFilters{
			Filters: bson.A{
				filterOfArray1,
			},
		},
	}

	col := c.collection(collection)
	return col.UpdateOne(ctx, filterOfDoc, update, &updateOpt)
}

func (c *client) getArrayItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, project bson.M, result interface{}) error {
	cond := make(bson.A, 0, len(filterOfArray))
	for k, v := range filterOfArray {
		cond = append(cond, bson.M{"$eq": bson.A{"$$this." + k, v}})
	}

	cond1 := cond[0]
	if len(filterOfArray) > 1 {
		cond1 = bson.M{"$and": cond}
	}

	filterOfArray1 := bson.M{
		array: bson.M{"$filter": bson.M{
			"input": fmt.Sprintf("$%s", array),
			"cond":  cond1,
		}},
	}

	for k, v := range project {
		if !strings.HasPrefix(k, array) {
			filterOfArray1[k] = v
		}
	}

	pipeline := bson.A{
		bson.M{"$match": filterOfDoc},
		bson.M{"$project": filterOfArray1},
		bson.M{"$project": project},
	}

	col := c.collection(collection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}
