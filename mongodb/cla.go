package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	col := c.db.Collection(claCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var v []CLA
	err = cursor.All(ctx, &v)
	if err != nil {
		return nil, err
	}

	r := make([]models.CLA, 0, len(v))
	for _, item := range v {
		cla := models.CLA{
			ID:        item.ID.String(),
			Name:      item.Name,
			Text:      item.Text,
			Language:  item.Language,
			Submitter: item.Submitter,
		}

		r = append(r, cla)
	}

	return r, nil
}
