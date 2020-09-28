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

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const clasCollection = "clas"

type CLA struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	Name      string             `bson:"name"`
	Text      string             `bson:"text"`
	Language  string             `bson:"language"`
	Submitter string             `bson:"submitter"`
	ApplyTo   string             `bson:"apply_to" required:"true"`
	Fields    []Field            `bson:"fields,omitempty"`
}

type Field struct {
	ID          string `bson:"id" required:"true"`
	Title       string `bson:"title"`
	Type        string `bson:"type"`
	Description string `bson:"description,omitempty"`
	Required    bool   `bson:"required"`
}

func (c *client) CreateCLA(cla dbmodels.CLA) (string, error) {
	body, err := golangsdk.BuildRequestBody(cla, "")
	if err != nil {
		return "", fmt.Errorf("build body failed, err:%v", err)
	}

	var r *mongo.UpdateResult

	f := func(ctx context.Context) error {
		col := c.collection(clasCollection)

		filter := bson.M{
			"name":      cla.Name,
			"submitter": cla.Submitter,
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
		return "", fmt.Errorf("the cla(%s) is already existing", cla.Name)
	}

	return toUID(r.UpsertedID)
}

func (this *client) DeleteCLA(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		col := this.collection(claOrgCollection)

		sr := col.FindOne(ctx, bson.M{"cla_id": uid})
		err := sr.Err()

		if err != nil {
			if isErrNoDocuments(err) {
				col = this.collection(clasCollection)

				_, err := col.DeleteOne(ctx, bson.M{"_id": oid})
				return err

			}
			return fmt.Errorf("failed to check whether the cla(%s) is bound: %v", uid, err)
		}

		return fmt.Errorf("can't delete the cla which has already been bound to org")

	}

	return this.doTransaction(f)
}

func (c *client) ListCLA(opts dbmodels.CLAListOptions) ([]dbmodels.CLA, error) {
	body, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, fmt.Errorf("build options to list cla failed, err:%v", err)
	}

	var v []CLA

	f := func(ctx context.Context) error {
		col := c.db.Collection(clasCollection)
		filter := bson.M(body)

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			return fmt.Errorf("error find clas: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLA: %v", err)
		}
		return nil
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	r := make([]dbmodels.CLA, 0, len(v))
	for _, item := range v {
		r = append(r, toModelCLA(item))
	}

	return r, nil
}

func (c *client) ListCLAByIDs(ids []string) ([]dbmodels.CLA, error) {
	ids1 := make(bson.A, 0, len(ids))
	for _, id := range ids {
		id1, err := toObjectID(id)
		if err != nil {
			return nil, err
		}
		ids1 = append(ids1, id1)
	}

	var v []CLA

	f := func(ctx context.Context) error {
		col := c.db.Collection(clasCollection)
		filter := bson.M{
			"_id": bson.M{"$in": ids1},
		}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			return fmt.Errorf("error find clas: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLA: %v", err)
		}
		return nil
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	r := make([]dbmodels.CLA, 0, len(v))
	for _, item := range v {
		r = append(r, toModelCLA(item))
	}

	return r, nil
}

func (c *client) GetCLA(uid string) (dbmodels.CLA, error) {
	var r dbmodels.CLA

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.db.Collection(clasCollection)
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

func toModelCLA(item CLA) dbmodels.CLA {
	cla := dbmodels.CLA{
		ID:        objectIDToUID(item.ID),
		Name:      item.Name,
		Text:      item.Text,
		Language:  item.Language,
		ApplyTo:   item.ApplyTo,
		Submitter: item.Submitter,
	}

	if item.Fields != nil {
		fs := make([]dbmodels.Field, 0, len(item.Fields))
		for _, v := range item.Fields {
			fs = append(fs, dbmodels.Field{
				ID:          v.ID,
				Title:       v.Title,
				Type:        v.Type,
				Description: v.Description,
				Required:    v.Required,
			})
		}
		cla.Fields = fs
	}

	return cla
}
