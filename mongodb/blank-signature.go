package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const blankSigCollection = "blank_signatures"

func (c *client) UploadBlankSignature(language string, pdf []byte) error {

	f := func(ctx context.Context) error {
		col := c.collection(blankSigCollection)

		insert := bson.M{
			"language": language,
			"pdf":      pdf,
		}

		upsert := true

		_, err := col.UpdateOne(
			ctx, bson.M{"language": language},
			bson.M{"$setOnInsert": insert},
			&options.UpdateOptions{Upsert: &upsert},
		)
		return err
	}

	return withContext(f)
}

func (c *client) DownloadBlankSignature(language string) ([]byte, error) {
	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.collection(blankSigCollection)

		sr = col.FindOne(ctx, bson.M{"language": language})
		return nil
	}

	withContext(f)

	var v struct {
		PDF []byte `bson:"pdf"`
	}

	err := sr.Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("error decoding to bson struct: %v", err)
	}

	return v.PDF, nil
}
