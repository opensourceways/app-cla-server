package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/zengchen1024/cla/models"
)

const collection = "clas"

type CLA struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	Name      string             `bson:"name"`
	Text      string             `bson:"text"`
	Language  string             `bson:"language"`
}

func (c *client) CreateCLA(cla models.CLA) (models.CLA, error) {
	body, err := golangsdk.BuildRequestBody(cla, "")
	if err != nil {
		return cla, fmt.Errorf("build body failed, err:%v", err)
	}

	col := c.db.Collection(collection)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

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
