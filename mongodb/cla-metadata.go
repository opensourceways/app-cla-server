package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zengchen1024/cla-server/models"
)

const claMetadataColl = "cla_metadatas"

type CLAMetadata struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	Name      string             `bson:"name"`
	Text      string             `bson:"text"`
	Language  string             `bson:"language"`
	Submitter string             `bson:"submitter"`
}

func (c *client) CreateCLAMetadata(data models.CLAMetadata) (string, error) {
	body, err := golangsdk.BuildRequestBody(data, "")
	if err != nil {
		return "", fmt.Errorf("build body failed, err:%v", err)
	}

	var r *mongo.UpdateResult

	f := func(ctx context.Context) error {
		col := c.collection(claMetadataColl)

		filter := bson.M{
			"name":      data.Name,
			"submitter": data.Submitter,
		}

		upsert := true

		r, err = col.UpdateOne(ctx, filter, bson.M{"$setOnInsert": bson.M(body)}, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return fmt.Errorf("write db failed, err:%v", err)
		}

		return nil
	}

	err = withContext(f)
	if err != nil {
		return "", err
	}

	if r.UpsertedID == nil {
		return "", fmt.Errorf("the cla metadata(%s) is already existing", data.Name)
	}

	return toUID(r.UpsertedID)
}

func (this *client) DeleteCLAMetadata(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		col := this.collection(orgRepoCollection)

		sr := col.FindOne(ctx, bson.M{"metadata_id": uid})
		err := sr.Err()

		if err != nil {
			if err.Error() == mongo.ErrNoDocuments.Error() {
				col = this.collection(claMetadataColl)

				_, err := col.DeleteOne(ctx, bson.M{"_id": oid})
				return err

			}
			return fmt.Errorf("failed to check whether the cla metadata(%s) is bound: %v", uid, err)
		}

		return fmt.Errorf("can't delete the cla metadata which has already been bound to org")

	}

	return this.doTransaction(f)
}

func (c *client) ListCLAMetadata(belongingTo []string) ([]models.CLAMetadata, error) {
	var v []CLAMetadata

	f := func(ctx context.Context) error {
		col := c.db.Collection(claMetadataColl)

		a := make(bson.A, 0, len(belongingTo))
		for _, v := range belongingTo {
			a = append(a, v)
		}

		filter := bson.M{
			"submitter": bson.M{"$in": a},
		}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			return fmt.Errorf("error find cla metadatas: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLAMetadata: %v", err)
		}
		return nil
	}

	err := withContext(f)
	if err != nil {
		return nil, err
	}

	r := make([]models.CLAMetadata, 0, len(v))
	for _, item := range v {
		r = append(r, toModelCLAMetadata(item))
	}

	return r, nil
}

func (c *client) GetCLAMetadata(uid string) (models.CLAMetadata, error) {
	var r models.CLAMetadata

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.db.Collection(claMetadataColl)
		sr = col.FindOne(ctx, bson.M{"_id": oid})
		return nil
	}

	withContext(f)

	var v CLAMetadata
	err = sr.Decode(&v)
	if err != nil {
		return r, fmt.Errorf("error decoding to bson struct of CLAMetadata: %v", err)
	}

	return toModelCLAMetadata(v), nil
}

func toModelCLAMetadata(item CLAMetadata) models.CLAMetadata {
	return models.CLAMetadata{
		ID:        objectIDToUID(item.ID),
		Name:      item.Name,
		Text:      item.Text,
		Language:  item.Language,
		Submitter: item.Submitter,
	}
}
