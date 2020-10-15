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

func structToMap(info interface{}) (bson.M, error) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     err,
		}
	}
	return bson.M(body), nil
}

func addCorporationID(email string, body bson.M) {
	body[fieldCorporationID] = genCorpID(email)
}

func genCorpID(email string) string {
	return util.EmailSuffix(email)
}

func filterOfCorpID(email string) bson.M {
	return bson.M{fieldCorporationID: genCorpID(email)}
}

func filterOfDocID(oid primitive.ObjectID) bson.M {
	return bson.M{"_id": oid}
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

func errorIfMatchingNoDoc(r *mongo.UpdateResult) error {
	if r.MatchedCount == 0 {
		return dbmodels.DBError{
			ErrCode: util.ErrNoDBRecord,
			Err:     fmt.Errorf("doesn't match any records"),
		}
	}
	return nil
}

func (c *client) pushArryItem(ctx context.Context, collection, array string, filterOfDoc, value bson.M) error {
	update := bson.M{"$push": bson.M{array: value}}

	col := c.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return err
	}

	return errorIfMatchingNoDoc(r)
}

func (c *client) pushArryItems(ctx context.Context, collection, array string, filterOfDoc bson.M, value bson.A) error {
	update := bson.M{"$push": bson.M{array: bson.M{"$each": value}}}

	col := c.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return err
	}

	return errorIfMatchingNoDoc(r)
}

func (c *client) pullArryItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) error {
	update := bson.M{"$pull": bson.M{array: filterOfArray}}

	col := c.collection(collection)
	r, err := col.UpdateOne(ctx, filterOfDoc, update)
	if err != nil {
		return err
	}

	return errorIfMatchingNoDoc(r)
}

// r, _ := col.UpdateOne; r.ModifiedCount == 0 will happen in two case: 1. no matched array item; 2 update repeatedly with same update cmd.
// checkModified = true when it can't exclude any case of above two; otherwise it can be set as false.
func (c *client) updateArryItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, updateCmd bson.M, checkModified bool) error {
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

	r, err := col.UpdateOne(ctx, filterOfDoc, update, &updateOpt)
	if err != nil {
		return err
	}

	if err := errorIfMatchingNoDoc(r); err != nil {
		return err
	}

	if r.ModifiedCount == 0 && checkModified {
		b, err := c.isArryItemNotExists(ctx, collection, array, filterOfDoc, filterOfArray)
		if err == nil && b {
			return dbmodels.DBError{
				ErrCode: util.ErrNoDBRecord,
				Err:     fmt.Errorf("can't find the corp signing record"),
			}
		}
	}
	return nil
}

func (c *client) isArryItemNotExists(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) (bool, error) {
	opts := options.FindOptions{
		Projection: bson.M{"_id": 1},
	}

	query := bson.M{array: bson.M{"$elemMatch": filterOfArray}}
	for k, v := range filterOfDoc {
		query[k] = v
	}

	col := c.collection(collection)

	cursor, err := col.Find(ctx, query, &opts)
	if err != nil {
		return false, err
	}

	var v []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err = cursor.All(ctx, &v); err != nil {
		return false, err
	}

	return len(v) <= 0, nil
}

func (c *client) getArrayItem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, project bson.M, result interface{}) error {
	pipeline := bson.A{bson.M{"$match": filterOfDoc}}

	if len(filterOfArray) > 0 {
		project1 := bson.M{
			array: bson.M{"$filter": bson.M{
				"input": fmt.Sprintf("$%s", array),
				"cond":  filterOfArrayItem(filterOfArray),
			}},
		}

		for k, v := range project {
			if !strings.HasPrefix(k, array) {
				project1[k] = v
			}
		}

		pipeline = append(pipeline, bson.M{"$project": project1})
	}

	if len(project) > 0 {
		pipeline = append(pipeline, bson.M{"$project": project})
	}

	col := c.collection(collection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}

func filterOfArrayItem(filterOfArray bson.M) bson.M {
	cond := make(bson.A, 0, len(filterOfArray))
	for k, v := range filterOfArray {
		cond = append(cond, bson.M{"$eq": bson.A{"$$this." + k, v}})
	}

	if len(filterOfArray) == 1 {
		return cond[0].(bson.M)
	}

	return bson.M{"$and": cond}
}
