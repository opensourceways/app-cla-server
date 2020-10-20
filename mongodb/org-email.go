package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const orgEmailCollection = "org_emails"

type OrgEmail struct {
	ID       primitive.ObjectID `bson:"_id" json:"-"`
	Email    string             `bson:"email" json:"email" required:"true"`
	Platform string             `bson:"platform" json:"platform" required:"true"`
	Token    []byte             `bson:"token" json:"-"`
}

func (c *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) error {
	info := OrgEmail{
		Email:    opt.Email,
		Platform: opt.Platform,
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}
	body["token"] = opt.Token

	f := func(ctx context.Context) error {
		_, err := c.newDocIfNotExist(ctx, orgEmailCollection, bson.M{"email": opt.Email}, body)
		if err != nil && isErrorOfRecordExists(err) {
			return nil
		}
		return err
	}

	return withContext(f)
}

func (c *client) GetOrgEmailInfo(email string) (dbmodels.OrgEmailCreateInfo, error) {
	var v OrgEmail

	f := func(ctx context.Context) error {
		return c.getDoc(ctx, orgEmailCollection, bson.M{"email": email}, bson.M{"email": 0}, &v)
	}

	if err := withContext(f); err != nil {
		return dbmodels.OrgEmailCreateInfo{}, err
	}

	return dbmodels.OrgEmailCreateInfo{
		Platform: v.Platform,
		Token:    v.Token,
	}, nil
}
