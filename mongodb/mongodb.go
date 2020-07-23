package mongodb

import (
	"context"
	"fmt"
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

func (this *client) Close() error {
	return withContext(this.c.Disconnect)
}

func (c *client) collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func (this *client) doTransaction(f func(mongo.SessionContext) error) error {

	callback := func(sc mongo.SessionContext) (interface{}, error) {
		return nil, f(sc)
	}

	s, err := this.c.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start mongodb session: %s", err.Error())
	}

	ctx := context.Background()
	defer s.EndSession(ctx)

	_, err = s.WithTransaction(ctx, callback)
	return err
}

func toObjectID(uid string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(uid)
}

func objectIDToUID(oid primitive.ObjectID) string {
	return oid.Hex()
}

func toUID(oid interface{}) (string, error) {
	v, ok := oid.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("retrieve id failed")
	}
	return v.Hex(), nil
}

func withContext(f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f(ctx)
}
