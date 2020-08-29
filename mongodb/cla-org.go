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

	"github.com/zengchen1024/cla-server/dbmodels"
)

const (
	claOrgCollection  = "cla_orgs"
	orgIdentifierName = "org_identifier"
)

func additionalConditionForCLAOrgDoc(filter bson.M) {
	filter["enabled"] = true
}

type signingInfo map[string]interface{}

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

	// Individuals is the cla signing information of ordinary contributors
	// key is the email of contributor
	Individuals map[string]signingInfo `bson:"individuals,omitempty"`

	// Employees is the cla signing information of employees and grouped by corporation
	// key is the email suffix of corporation
	Employees map[string][]employeeSigning `bson:"employees,omitempty"`

	// Corporations is the cla signing information of corporation
	// key is the email suffix of corporation
	Corporations []corporationSigning `bson:"corporations,omitempty"`

	// CorporationManagers is the managers of corporation who can manage the employee
	CorporationManagers []corporationManager `bson:"corporation_managers,omitempty"`
}

func orgIdentifier(platform, org string) string {
	return fmt.Sprintf("%s:%s", platform, org)
}

func (c *client) CreateBindingBetweenCLAAndOrg(claOrg dbmodels.CLAOrg) (string, error) {
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

func (c *client) DeleteBindingBetweenCLAAndOrg(uid string) error {
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

func (c *client) GetBindingBetweenCLAAndOrg(uid string) (dbmodels.CLAOrg, error) {
	var r dbmodels.CLAOrg

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var sr *mongo.SingleResult

	f := func(ctx context.Context) error {
		col := c.db.Collection(claOrgCollection)
		opt := options.FindOneOptions{
			Projection: bson.M{
				"individuals":          0,
				"employees":            0,
				"corporations":         0,
				"corporation_managers": 0,
			},
		}

		sr = col.FindOne(ctx, bson.M{"_id": oid}, &opt)
		return nil
	}

	withContext(f)

	var v CLAOrg
	err = sr.Decode(&v)
	if err != nil {
		return r, fmt.Errorf("error decoding to bson struct of CLA: %v", err)
	}

	return toModelCLAOrg(v), nil
}

func (c *client) ListBindingBetweenCLAAndOrg(opt dbmodels.CLAOrgListOption) ([]dbmodels.CLAOrg, error) {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return nil, fmt.Errorf("build options to list cla-org failed, err:%v", err)
	}
	filter := bson.M(body)
	additionalConditionForCLAOrgDoc(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.db.Collection(claOrgCollection)

		opts := options.FindOptions{
			Projection: bson.M{
				"individuals":          0,
				"employees":            0,
				"corporations":         0,
				"corporation_managers": 0,
			},
		}
		cursor, err := col.Find(ctx, filter, &opts)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of CLAOrg: %v", err)
		}
		return nil
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	n := len(v)
	r := make([]dbmodels.CLAOrg, 0, n)
	for _, item := range v {
		r = append(r, toModelCLAOrg(item))
	}

	if opt.RepoID != "" {
		r1 := make([]dbmodels.CLAOrg, 0, n)
		for i := 0; i < n; i++ {
			if r[i].RepoID == opt.RepoID {
				r1 = append(r1, r[i])
			}
		}
		if len(r1) != 0 {
			return r1, nil
		}
	}
	return r, nil
}

func toModelCLAOrg(item CLAOrg) dbmodels.CLAOrg {
	return dbmodels.CLAOrg{
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
	}
}
