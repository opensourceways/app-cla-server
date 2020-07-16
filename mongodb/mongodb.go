package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zengchen1024/cla-server/models"
)

var _ models.IDB = (*client)(nil)

type client struct {
	c  *mongo.Client
	db *mongo.Database
}

func RegisterDatabase(conn, db string) (*client, error) {
	c, err := mongo.NewClient(options.Client().ApplyURI(conn))
	if err != nil {
		return nil, err
	}
	err = withContext(c.Connect)
	if err != nil {
		return nil, err
	}

	cli := &client{
		c:  c,
		db: c.Database(db),
	}
	return cli, nil
}

func (c *client) collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func toObjectID(uid string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(uid)
}

func toUID(oid primitive.ObjectID) string {
	return oid.Hex()
}

func withContext(f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}
