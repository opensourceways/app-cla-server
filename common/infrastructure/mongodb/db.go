package mongodb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/opensourceways/app-cla-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cli *client

func Init(cfg *Config) error {
	rootPEM, err := ioutil.ReadFile(cfg.CAFile)
	err1 := os.Remove(cfg.CAFile)
	if err2 := util.MultiErrors(err, err1); err2 != nil {
		return err2
	}

	roots := x509.NewCertPool()

	if ok := roots.AppendCertsFromPEM([]byte(rootPEM)); !ok {
		return fmt.Errorf("fail to get certs from %s", cfg.CAFile)
	}

	tlsConfig := &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
	}

	clientOpts := options.Client().ApplyURI(cfg.Conn)
	clientOpts.SetTLSConfig(tlsConfig)

	c, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return fmt.Errorf("connect err: %s", err.Error())
	}

	timeout := cfg.timeout()

	// verify if database connection is created successfully
	err = withContext(
		func(ctx context.Context) error {
			return c.Ping(ctx, nil)
		},
		timeout,
	)
	if err != nil {
		return fmt.Errorf("ping err: %s", err.Error())
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
