package mongodb

import (
	"context"
	"time"

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
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	cli := &client{
		c:  c,
		db: c.Database(db),
	}
	return cli, nil
}
