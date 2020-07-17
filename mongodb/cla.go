package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zengchen1024/cla-server/models"
)

const claCollection = "clas"

type CLA struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	Name      string             `bson:"name"`
	Text      string             `bson:"text"`
	Language  string             `bson:"language"`
	Submitter string             `bson:"submitter"`
}

func (c *client) CreateCLA(cla models.CLA) (models.CLA, error) {
	body, err := golangsdk.BuildRequestBody(cla, "")
	if err != nil {
		return cla, fmt.Errorf("build body failed, err:%v", err)
	}

	col := c.db.Collection(claCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := col.InsertOne(ctx, bson.M(body))
	if err != nil {
		return cla, fmt.Errorf("write db failed, err:%v", err)
	}

	v, ok := r.InsertedID.(primitive.ObjectID)
	if !ok {
		return cla, fmt.Errorf("retrieve id failed")
	}

	cla.ID = v.String()
	return cla, nil
}

func (c *client) ListCLA() ([]models.CLA, error) {
	var v []CLA

	f := func(ctx context.Context) error {
		col := c.db.Collection(claCollection)

		cursor, err := col.Find(ctx, bson.D{})
		if err != nil {
			return fmt.Errorf("error find clas: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLA: %v", err)
		}
		return nil
	}

	err := withContext(f)
	if err != nil {
		return nil, err
	}

	r := make([]models.CLA, 0, len(v))
	for _, item := range v {
		r = append(r, toModelCLA(item))
	}

	return r, nil
}

func (c *client) GetCLA(uid string) (models.CLA, error) {
	var r models.CLA

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.db.Collection(claCollection)
		sr = col.FindOne(ctx, bson.M{"_id": oid})
		return nil
	}

	withContext(f)

	var v CLA
	err = sr.Decode(&v)
	if err != nil {
		return r, fmt.Errorf("error decoding to bson struct of CLA: %v", err)
	}

	return toModelCLA(v), nil
}

func toModelCLA(item CLA) models.CLA {
	return models.CLA{
		ID:        objectIDToUID(item.ID),
		Name:      item.Name,
		Text:      item.Text,
		Language:  item.Language,
		Submitter: item.Submitter,
	}
}
