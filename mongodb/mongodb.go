package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
)

var _ dbmodels.IDB = (*client)(nil)

type collection interface {
	Collection(name string) *mongo.Collection
}

type client struct {
	col collection

	vcCollection                string
	orgEmailCollection          string
	corpPDFCollection           string
	claPDFCollection            string
	linkCollection              string
	corpSigningCollection       string
	individualSigningCollection string
}

func Initialize(col collection, cfg *config.MongodbConfig) *client {
	return &client{
		col: col,

		vcCollection:                cfg.VCCollection,
		orgEmailCollection:          cfg.OrgEmailCollection,
		corpPDFCollection:           cfg.CorpPDFCollection,
		claPDFCollection:            cfg.CLAPDFCollection,
		linkCollection:              cfg.LinkCollection,
		corpSigningCollection:       cfg.CorpSigningCollection,
		individualSigningCollection: cfg.IndividualSigningCollection,
	}
}

func (c *client) collection(name string) *mongo.Collection {
	return c.col.Collection(name)
}

func (c *client) Close() error {
	return nil
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
