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

const orgRepoCollection = "org_repos"

type OrgRepo struct {
	ID primitive.ObjectID `bson:"_id"`

	CreatedAt   time.Time `bson:"created_at,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at,omitempty"`
	Platform    string    `bson:"platform"`
	OrgID       string    `bson:"org_id"`
	RepoID      string    `bson:"repo_id"`
	CLAID       string    `bson:"cla_id"`
	CLALanguage string    `bson:"cla_language"`
	MetadataID  string    `bson:"metadata_id,omitempty"`
	OrgEmail    string    `bson:"org_email,omitempty"`
	Enabled     bool      `bson:"enabled"`
	Submitter   string    `bson:"submitter"`
}

func repoIdentifier(platform, org, repo string) string {
	if repo == "" {
		return fmt.Sprintf("%s:%s", platform, org)
	}
	return fmt.Sprintf("%s:%s:%s", platform, org, repo)
}

func (c *client) CreateOrgRepo(orgRepo models.OrgRepo) (string, error) {
	body, err := golangsdk.BuildRequestBody(orgRepo, "")
	if err != nil {
		return "", fmt.Errorf("build body failed, err:%v", err)
	}
	body["repo_identifier"] = repoIdentifier(orgRepo.Platform, orgRepo.OrgID, orgRepo.RepoID)

	var r *mongo.UpdateResult

	f := func(ctx context.Context) error {
		col := c.collection(orgRepoCollection)

		filter := bson.M{
			"platform":     orgRepo.Platform,
			"org_id":       orgRepo.OrgID,
			"repo_id":      orgRepo.RepoID,
			"cla_language": orgRepo.CLALanguage,
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
			orgRepo.Platform, orgRepo.OrgID, orgRepo.RepoID, orgRepo.CLALanguage)
	}

	return toUID(r.UpsertedID)
}

func (c *client) DisableOrgRepo(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(orgRepoCollection)

		_, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"enabled": false}})
		return err
	}

	return withContext(f)
}

func (c *client) ListOrgRepo(opt models.OrgRepos) ([]models.OrgRepo, error) {
	var v []OrgRepo

	f := func(ctx context.Context) error {
		col := c.db.Collection(orgRepoCollection)

		var ids bson.A
		for platform, orgs := range opt.Org {
			for _, org := range orgs {
				ids = append(ids, fmt.Sprintf("/%s.*/", repoIdentifier(platform, org, "")))
			}
		}

		filter := bson.M{"repo_identifier": bson.M{"$in": ids}}

		cursor, err := col.Find(ctx, filter)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of OrgRepo: %v", err)
		}
		return nil
	}

	err := withContext(f)
	if err != nil {
		return nil, err
	}

	r := make([]models.OrgRepo, 0, len(v))
	for _, item := range v {
		r = append(r, toModelOrgRepo(item))
	}

	return r, nil
}

func toModelOrgRepo(item OrgRepo) models.OrgRepo {
	return models.OrgRepo{
		ID:          objectIDToUID(item.ID),
		Platform:    item.Platform,
		OrgID:       item.OrgID,
		RepoID:      item.RepoID,
		CLAID:       item.CLAID,
		CLALanguage: item.CLALanguage,
		MetadataID:  item.MetadataID,
		OrgEmail:    item.OrgEmail,
		Enabled:     item.Enabled,
		Submitter:   item.Submitter,
	}
}
