package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const clasCollection = "clas"

type CLA struct {
	ID        primitive.ObjectID `bson:"_id" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"-"`
	UpdatedAt time.Time          `bson:"updated_at" json:"-"`
	URL       string             `bson:"url" json:"url" required:"true"`
	Text      string             `bson:"text" json:"text" required:"true"`
	Language  string             `bson:"language" json:"language" required:"true"`
	Fields    []Field            `bson:"fields" json:"fields,omitempty"`
}

type Field struct {
	ID          string `bson:"id" json:"id" required:"true"`
	Title       string `bson:"title" json:"title" required:"true"`
	Type        string `bson:"type" json:"type" required:"true"`
	Description string `bson:"description" json:"description,omitempty"`
	Required    bool   `bson:"required" json:"required"`
}

func (c *client) CreateCLA(cla dbmodels.CLA) (string, error) {
	info := CLA{
		URL:      cla.Name,
		Text:     cla.Text,
		Language: cla.Language,
	}
	if len(cla.Fields) > 0 {
		fields := make([]Field, 0, len(cla.Fields))
		for _, item := range cla.Fields {
			fields = append(fields, Field{
				ID:          item.ID,
				Title:       item.Title,
				Type:        item.Type,
				Description: item.Description,
				Required:    item.Required,
			})
		}
	}
	body, err := structToMap(info)
	if err != nil {
		return "", err
	}

	uid := ""
	f := func(ctx context.Context) error {
		s, err := c.insertDoc(ctx, clasCollection, body)
		uid = s
		return err
	}

	err = withContext(f)
	return uid, err
}

func (this *client) DeleteCLA(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		col := this.collection(orgCLACollection)

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
		filter := bson.M{
			"_id": bson.M{"$in": ids1},
		}

		return c.getDocs(ctx, clasCollection, filter, nil, &v)
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

func (c *client) GetCLA(uid string, onlyFields bool) (dbmodels.CLA, error) {
	oid, err := toObjectID(uid)
	if err != nil {
		return dbmodels.CLA{}, err
	}

	var v CLA

	project := bson.M{}
	if onlyFields {
		project["fields"] = 1
	}
	f := func(ctx context.Context) error {
		return c.getDoc(ctx, clasCollection, filterOfDocID(oid), project, &v)
	}

	if err := withContext(f); err != nil {
		return dbmodels.CLA{}, err
	}

	return toModelCLA(v), nil
}

func toModelCLA(item CLA) dbmodels.CLA {
	cla := dbmodels.CLA{
		ID:       objectIDToUID(item.ID),
		Name:     item.URL,
		Text:     item.Text,
		Language: item.Language,
	}

	if len(item.Fields) > 0 {
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
