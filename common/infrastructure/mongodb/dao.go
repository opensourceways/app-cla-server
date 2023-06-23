package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
