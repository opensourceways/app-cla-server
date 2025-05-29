package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	fieldIndex   = "_id"
	fieldVersion = "version"

	mongoCmdIn          = "$in"
	mongoCmdAll         = "$all"
	mongoCmdSet         = "$set"
	mongoCmdInc         = "$inc"
	mongoCmdEach        = "$each"
	mongoCmdPush        = "$push"
	mongoCmdPull        = "$pull"
	mongoCmdMatch       = "$match"
	mongoCmdFilter      = "$filter"
	mongoCmdProject     = "$project"
	mongoCmdAddToSet    = "$addToSet"
	mongoCmdElemMatch   = "$elemMatch"
	mongoCmdSetOnInsert = "$setOnInsert"
)

var (
	errDocExists    = errors.New("doc exists")
	errDocNotExists = errors.New("doc doesn't exist")
)

func isErrOfNoDocuments(err error) bool {
	return err.Error() == mongo.ErrNoDocuments.Error()
}

type daoImpl struct {
	col     *mongo.Collection
	timeout time.Duration
}

func (impl *daoImpl) withContext(f func(context.Context) error) error {
	return withContext(f, impl.timeout)
}

func (impl *daoImpl) IsDocNotExists(err error) bool {
	return errors.Is(err, errDocNotExists)
}

func (impl *daoImpl) IsDocExists(err error) bool {
	return errors.Is(err, errDocExists)
}

func (impl *daoImpl) InsertDoc(doc bson.M) (string, error) {
	docId := ""

	err := impl.withContext(func(ctx context.Context) error {
		r, err := impl.col.InsertOne(ctx, doc)
		if err != nil {
			return err
		}

		docId = toDocId(r.InsertedID)

		return nil
	})

	return docId, err
}

func (impl *daoImpl) InsertDocIfNotExists(filter, doc bson.M) (string, error) {
	docId := ""

	err := impl.withContext(func(ctx context.Context) error {
		upsert := true

		r, err := impl.col.UpdateOne(
			ctx, filter, bson.M{"$setOnInsert": doc},
			&options.UpdateOptions{Upsert: &upsert},
		)
		if err != nil {
			return err
		}

		if r.UpsertedID == nil {
			return errDocExists
		}

		docId = toDocId(r.UpsertedID)

		return nil
	})

	return docId, err
}

func (impl *daoImpl) PushArraySingleItemAndUpdate(filter bson.M, array string, v interface{}, u bson.M, version int) error {
	return impl.updateDoc(
		filter, version, bson.M{
			mongoCmdPush: bson.M{array: v},
			mongoCmdSet:  u,
		},
	)
}

func (impl *daoImpl) PushArraySingleItem(filter bson.M, array string, v interface{}, version int) error {
	return impl.updateDoc(
		filter, version, bson.M{mongoCmdPush: bson.M{array: v}},
	)
}

func (impl *daoImpl) PushArrayMultiItems(filter bson.M, array string, value bson.A, version int) error {
	return impl.updateDoc(
		filter, version,
		bson.M{mongoCmdPush: bson.M{array: bson.M{mongoCmdEach: value}}},
	)
}

func (impl *daoImpl) PullArrayMultiItems(filter bson.M, array string, filterOfItem bson.M, version int) error {
	return impl.updateDoc(
		filter, version,
		bson.M{mongoCmdPull: bson.M{array: filterOfItem}},
	)
}

func (impl *daoImpl) MoveArrayItem(filter bson.M, from string, filterOfItem bson.M, to string, value bson.M, version int) error {
	return impl.updateDoc(
		filter, version,
		bson.M{
			mongoCmdPull: bson.M{from: filterOfItem},
			mongoCmdPush: bson.M{to: value},
		},
	)
}

func (impl *daoImpl) UpdateDocsWithoutVersion(filter bson.M, v bson.M) error {
	return impl.withContext(func(ctx context.Context) error {
		_, err := impl.col.UpdateMany(ctx, filter, bson.M{mongoCmdSet: v})

		return err
	})
}

func (impl *daoImpl) UpdateDoc(filter bson.M, v bson.M, version int) error {
	return impl.updateDoc(filter, version, bson.M{mongoCmdSet: v})
}

func (impl *daoImpl) ReplaceDoc(filter, doc bson.M) (string, error) {
	docId := ""

	err := impl.withContext(func(ctx context.Context) error {
		upsert := true

		r, err := impl.col.ReplaceOne(
			ctx, filter, doc,
			&options.ReplaceOptions{Upsert: &upsert},
		)
		if err != nil {
			return err
		}

		if r.UpsertedID != nil {
			docId = toDocId(r.UpsertedID)
		}

		return nil
	})

	return docId, err
}

func (impl *daoImpl) updateDoc(filter bson.M, version int, cmd bson.M) error {
	return impl.withContext(func(ctx context.Context) error {
		filter[fieldVersion] = version
		cmd[mongoCmdInc] = bson.M{fieldVersion: 1}

		r, err := impl.col.UpdateOne(ctx, filter, cmd)

		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return errDocNotExists
		}

		return nil
	})
}

func (impl *daoImpl) UpdateArraySingleItem(filter bson.M, array string, filterOfArray, doc bson.M, version int) error {
	return impl.withContext(func(ctx context.Context) error {
		filter[fieldVersion] = version

		cmd := bson.M{}
		for k, v := range doc {
			cmd[fmt.Sprintf("%s.$[i].%s", array, k)] = v
		}

		arrayFilter := bson.M{}
		for k, v := range filterOfArray {
			arrayFilter["i."+k] = v
		}

		r, err := impl.col.UpdateOne(
			ctx, filter,
			bson.M{
				mongoCmdSet: cmd,
				mongoCmdInc: bson.M{fieldVersion: 1},
			},
			&options.UpdateOptions{
				ArrayFilters: &options.ArrayFilters{
					Filters: bson.A{
						arrayFilter,
					},
				},
			},
		)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return errDocNotExists
		}

		return nil
	})
}

func (impl *daoImpl) GetDoc(filter, project bson.M, result interface{}) error {
	return impl.withContext(func(ctx context.Context) error {
		var sr *mongo.SingleResult

		if len(project) > 0 {
			sr = impl.col.FindOne(ctx, filter, &options.FindOneOptions{
				Projection: project,
			})
		} else {
			sr = impl.col.FindOne(ctx, filter)
		}

		err := sr.Decode(result)
		if err != nil && isErrOfNoDocuments(err) {
			return errDocNotExists
		}

		return err
	})
}

func (impl *daoImpl) GetDocs(filter, project bson.M, result interface{}) error {
	return impl.withContext(func(ctx context.Context) error {
		var cursor *mongo.Cursor
		var err error

		if len(project) > 0 {
			cursor, err = impl.col.Find(ctx, filter, &options.FindOptions{
				Projection: project,
			})
		} else {
			cursor, err = impl.col.Find(ctx, filter)
		}

		if err != nil {
			return err
		}

		return cursor.All(ctx, result)
	})
}

func (impl *daoImpl) GetDocAndDelete(filter, project bson.M, result interface{}) error {
	return impl.withContext(func(ctx context.Context) error {
		var sr *mongo.SingleResult

		if len(project) > 0 {
			sr = impl.col.FindOneAndDelete(
				ctx, filter,
				&options.FindOneAndDeleteOptions{
					Projection: project,
				},
			)
		} else {
			sr = impl.col.FindOneAndDelete(ctx, filter)
		}

		err := sr.Decode(result)
		if err != nil && isErrOfNoDocuments(err) {
			return errDocNotExists
		}

		return err
	})
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

func (impl *daoImpl) GetArrayItem(
	filter bson.M, array string, filterOfArray, project bson.M, result interface{},
) error {
	return impl.withContext(func(ctx context.Context) error {
		pipeline := bson.A{bson.M{"$match": filter}}

		project1 := bson.M{}

		project1[array] = bson.M{"$filter": bson.M{
			"input": fmt.Sprintf("$%s", array),
			"cond":  conditionTofilterArray(filterOfArray),
		}}

		pipeline = append(pipeline, bson.M{"$project": project1})

		if len(project) > 0 {
			pipeline = append(pipeline, bson.M{"$project": project})
		}

		cursor, err := impl.col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, result)
	})
}

func (impl *daoImpl) DeleteDoc(filter bson.M) error {
	return impl.withContext(func(ctx context.Context) error {
		r, err := impl.col.DeleteOne(ctx, filter)
		if err != nil {
			return err
		}

		if r.DeletedCount == 0 {
			return errDocNotExists
		}

		return nil
	})
}

func (impl *daoImpl) DeleteDocs(filter bson.M) error {
	return impl.withContext(func(ctx context.Context) error {
		_, err := impl.col.DeleteMany(ctx, filter)

		return err
	})
}

func (impl *daoImpl) NewDocId() string {
	return primitive.NewObjectID().Hex()
}

func (impl *daoImpl) DocIdFilter(s string) (bson.M, error) {
	v, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return nil, err
	}

	return bson.M{
		fieldIndex: v,
	}, nil
}

func (impl *daoImpl) DocIdsFilter(ids []string) (bson.M, error) {
	oids := make(bson.A, len(ids))

	for i, s := range ids {
		v, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			return nil, err
		}

		oids[i] = v
	}

	return bson.M{
		fieldIndex: bson.M{mongoCmdIn: oids},
	}, nil
}
