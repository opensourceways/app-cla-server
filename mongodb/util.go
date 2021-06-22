package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/util"
)

func genCorpID(email string) string {
	return util.EmailSuffix(email)
}

func filterOfCorpID(email string) bson.M {
	return bson.M{fieldCorpID: genCorpID(email)}
}

func isErrNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

func (this *client) isArrayElemNotExists(ctx context.Context, collection, array string, filterOfDoc, filterOfArray bson.M) (bool, error) {
	query := bson.M{array: bson.M{"$elemMatch": filterOfArray}}
	for k, v := range filterOfDoc {
		query[k] = v
	}

	var v []struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := this.getDocs(ctx, collection, query, bson.M{"_id": 1}, &v)
	if err != nil {
		return false, err
	}

	return len(v) <= 0, nil
}

func (this *client) getArrayElem(ctx context.Context, collection, array string, filterOfDoc, filterOfArray, project bson.M, result interface{}) error {
	ma := map[string]bson.M{}
	if len(filterOfArray) > 0 {
		ma[array] = filterOfArray
	}
	return this.getMultiArrays(ctx, collection, filterOfDoc, ma, project, result)
}

func (this *client) getMultiArrays(ctx context.Context, collection string, filterOfDoc bson.M, filterOfArrays map[string]bson.M, project bson.M, result interface{}) error {
	m := map[string]func() bson.M{}
	for k, v := range filterOfArrays {
		m[k] = func() bson.M {
			return conditionTofilterArray(v)
		}
	}

	return this.getArrayElems(ctx, collection, filterOfDoc, project, m, result)
}

func (this *client) getArrayElems(ctx context.Context, collection string, filterOfDoc bson.M, project bson.M, filterOfArrays map[string]func() bson.M, result interface{}) error {
	pipeline := bson.A{bson.M{"$match": filterOfDoc}}

	if len(filterOfArrays) > 0 {
		project1 := bson.M{}

		for array, cond := range filterOfArrays {
			project1[array] = bson.M{"$filter": bson.M{
				"input": fmt.Sprintf("$%s", array),
				"cond":  cond(),
			}}
		}

		for k, v := range project {
			s := k
			if i := strings.Index(k, "."); i >= 0 {
				s = k[:i]
			}
			if _, ok := filterOfArrays[s]; !ok {
				project1[k] = v
			}
		}

		pipeline = append(pipeline, bson.M{"$project": project1})
	}

	if len(project) > 0 {
		pipeline = append(pipeline, bson.M{"$project": project})
	}

	col := this.collection(collection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}

func conditionTofilterArray(filterOfArray bson.M) bson.M {
	cond := make(bson.A, 0, len(filterOfArray))
	for k, v := range filterOfArray {
		cond = append(cond, bson.M{"$eq": bson.A{"$$this." + k, v}})
	}

	if len(filterOfArray) == 1 {
		return cond[0].(bson.M)
	}

	return bson.M{"$and": cond}
}

func arrayElemFilter(array string, filterOfArray bson.M) bson.M {
	return bson.M{"$filter": bson.M{
		"input": fmt.Sprintf("$%s", array),
		"cond":  conditionTofilterArray(filterOfArray),
	}}
}

func (this *client) getDocs(ctx context.Context, collection string, filterOfDoc, project bson.M, result interface{}) error {
	col := this.collection(collection)

	var cursor *mongo.Cursor
	var err error
	if len(project) > 0 {
		cursor, err = col.Find(ctx, filterOfDoc, &options.FindOptions{
			Projection: project,
		})
	} else {
		cursor, err = col.Find(ctx, filterOfDoc)
	}

	if err != nil {
		return err
	}
	return cursor.All(ctx, result)
}

func (this *client) insertDoc(ctx context.Context, collection string, docInfo bson.M) (string, error) {
	col := this.collection(collection)
	r, err := col.InsertOne(ctx, docInfo)
	if err != nil {
		return "", err
	}

	return toUID(r.InsertedID)
}
