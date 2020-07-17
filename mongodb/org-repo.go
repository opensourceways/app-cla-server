package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zengchen1024/cla-server/models"
)

const orgRepoCollection = "org_repos"

func (c *client) CreateOrgRepo(orgRepo models.OrgRepo) (string, error) {
	body, err := golangsdk.BuildRequestBody(orgRepo, "")
	if err != nil {
		return "", fmt.Errorf("build body failed, err:%v", err)
	}

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
