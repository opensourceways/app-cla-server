package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *client) UploadOrgSignature(claOrgID string, pdf []byte) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		v := bson.M{
			fieldOrgSignature:    pdf,
			fieldOrgSignatureTag: true,
		}
		_, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": v})
		return err
	}

	return withContext(f)
}

func (c *client) DownloadOrgSignature(claOrgID string) ([]byte, error) {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return nil, err
	}

	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		opt := options.FindOneOptions{
			Projection: bson.M{
				fieldOrgSignature: 1,
			},
		}

		sr = col.FindOne(ctx, bson.M{"_id": oid, fieldOrgSignatureTag: true}, &opt)
		return nil
	}

	withContext(f)

	var v CLAOrg
	err = sr.Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("error decoding to bson struct of CLAOrg: %v", err)
	}

	return v.OrgSignature, nil
}
