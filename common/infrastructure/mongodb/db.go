package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cli *client

func Init(cfg *Config) error {
	c, err := mongo.NewClient(options.Client().ApplyURI(cfg.Conn))
	if err != nil {
		return err
	}

	timeout := cfg.timeout()

	if err = withContext(c.Connect, timeout); err != nil {
		return err
	}

	// verify if database connection is created successfully
	err = withContext(
		func(ctx context.Context) error {
			return c.Ping(ctx, nil)
		},
		timeout,
	)
	if err != nil {
		return err
	}

	cli = &client{
		c:       c,
		db:      c.Database(cfg.DBName),
		timeout: timeout,
	}

	return nil
}

func Close() error {
	if cli != nil {
		return cli.disconnect()
	}

	return nil
}

func DAO(name string) *daoImpl {
	return &daoImpl{
		col:     cli.Collection(name),
		timeout: cli.timeout,
	}
}

func Collection() *client {
	return cli
}

func withContext(f func(context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return f(ctx)
}

func toDocId(oid interface{}) string {
	if v, ok := oid.(primitive.ObjectID); ok {
		return v.Hex()
	}

	return ""
}

// client
type client struct {
	c       *mongo.Client
	db      *mongo.Database
	timeout time.Duration
}

func (cli *client) disconnect() error {
	return withContext(cli.c.Disconnect, cli.timeout)
}

func (cli *client) Collection(name string) *mongo.Collection {
	return cli.db.Collection(name)
}
