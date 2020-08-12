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

const (
	claOrgCollection  = "cla_orgs"
	orgIdentifierName = "org_identifier"
)

type CLAOrg struct {
	ID primitive.ObjectID `bson:"_id"`

	CreatedAt   time.Time `bson:"created_at,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at,omitempty"`
	Platform    string    `bson:"platform"`
	OrgID       string    `bson:"org_id"`
	RepoID      string    `bson:"repo_id"`
	CLAID       string    `bson:"cla_id"`
	CLALanguage string    `bson:"cla_language"`
	ApplyTo     string    `bson:"apply_to" required:"true"`
	OrgEmail    string    `bson:"org_email,omitempty"`
	Enabled     bool      `bson:"enabled"`
	Submitter   string    `bson:"submitter"`

	// Individuals is the cla signed information of ordinary contributors
	// key is the email of contributor
	Individuals map[string]models.IndividualSigning `bson:"individuals,omitempty"`
}

func orgIdentifier(platform, org string) string {
	return fmt.Sprintf("%s:%s", platform, org)
}

func (c *client) BindCLAToOrg(claOrg models.CLAOrg) (string, error) {
	body, err := golangsdk.BuildRequestBody(claOrg, "")
	if err != nil {
		return "", fmt.Errorf("build body failed, err:%v", err)
	}
	body[orgIdentifierName] = orgIdentifier(claOrg.Platform, claOrg.OrgID)

	var r *mongo.UpdateResult

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{
			"platform":     claOrg.Platform,
			"org_id":       claOrg.OrgID,
			"repo_id":      claOrg.RepoID,
			"cla_language": claOrg.CLALanguage,
			"apply_to":     claOrg.ApplyTo,
			"enabled":      true,
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
		return "", fmt.Errorf("the org/repo:%s/%s/%s has already been bound a cla with language:%s",
			claOrg.Platform, claOrg.OrgID, claOrg.RepoID, claOrg.CLALanguage)
	}

	return toUID(r.UpsertedID)
}

func (c *client) UnbindCLAFromOrg(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		v := bson.M{"enabled": false, "updated_at": time.Now()}
		_, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": v})
		return err
	}

	return withContext(f)
}

func (c *client) ListBindingOfCLAAndOrg(opt models.CLAOrgs) ([]models.CLAOrg, error) {
	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.db.Collection(claOrgCollection)

		var ids bson.A
		for platform, orgs := range opt.Org {
			for _, org := range orgs {
				ids = append(ids, orgIdentifier(platform, org))
			}
		}

		filter := bson.M{orgIdentifierName: bson.M{"$in": ids}}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLAOrg: %v", err)
		}
		return nil
	}

	err := withContext(f)
	if err != nil {
		return nil, err
	}

	r := make([]models.CLAOrg, 0, len(v))
	for _, item := range v {
		r = append(r, toModelCLAOrg(item))
	}

	return r, nil
}

func toModelCLAOrg(item CLAOrg) models.CLAOrg {
	return models.CLAOrg{
		ID:          objectIDToUID(item.ID),
		Platform:    item.Platform,
		OrgID:       item.OrgID,
		RepoID:      item.RepoID,
		CLAID:       item.CLAID,
		CLALanguage: item.CLALanguage,
		ApplyTo:     item.ApplyTo,
		OrgEmail:    item.OrgEmail,
		Enabled:     item.Enabled,
		Submitter:   item.Submitter,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
