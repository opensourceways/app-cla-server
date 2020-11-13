package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type OrgEmail struct {
	ID       primitive.ObjectID `bson:"_id" json:"-"`
	Email    string             `bson:"email" json:"email" required:"true"`
	Platform string             `bson:"platform" json:"platform" required:"true"`
	Token    []byte             `bson:"token" json:"-"`
}

func (this *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) error {
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
		_, err := this.newDocIfNotExist(ctx, this.orgEmailCollection, bson.M{"email": opt.Email}, body)
		if err != nil && isErrorOfRecordExists(err) {
			return nil
		}
		return err
	}

	return withContext(f)
}

func (this *client) GetOrgEmailInfo(email string) (dbmodels.OrgEmailCreateInfo, error) {
	var v OrgEmail

	f := func(ctx context.Context) error {
		return this.getDoc(ctx, this.orgEmailCollection, bson.M{"email": email}, bson.M{"email": 0}, &v)
	}

	if err := withContext(f); err != nil {
		return dbmodels.OrgEmailCreateInfo{}, err
	}

	return dbmodels.OrgEmailCreateInfo{
		Platform: v.Platform,
		Token:    v.Token,
	}, nil
}
