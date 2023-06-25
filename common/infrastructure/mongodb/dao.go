package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	fieldVersion = "version"

	mongoCmdAll         = "$all"
	mongoCmdSet         = "$set"
	mongoCmdInc         = "$inc"
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

func (impl *daoImpl) PushArrayDoc(filter bson.M, array string, v bson.M, version int) error {
	return impl.updateDoc(filter, bson.M{array: v}, version, mongoCmdPush)
}

func (impl *daoImpl) UpdateDoc(filter bson.M, field string, v bson.M, version int) error {
	return impl.updateDoc(filter, bson.M{field: v}, version, mongoCmdSet)
}

func (impl *daoImpl) updateDoc(filter, doc bson.M, version int, cmd string) error {
	return impl.withContext(func(ctx context.Context) error {
		filter[fieldVersion] = version

		r, err := impl.col.UpdateOne(
			ctx, filter,
			bson.M{
				cmd:         doc,
				mongoCmdInc: bson.M{fieldVersion: 1},
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

func (impl *daoImpl) NewDocId() string {
	return primitive.NewObjectID().Hex()
}

func (impl *daoImpl) DocIdFilter(s string) (bson.M, error) {
	v, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return nil, err
	}

	return bson.M{
		"_id": v,
	}, nil
}
