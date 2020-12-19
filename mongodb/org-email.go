package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) *dbmodels.DBError {
	info := cOrgEmail{
		Email:    opt.Email,
		Platform: opt.Platform,
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}
	body["token"] = opt.Token

	f := func(ctx context.Context) *dbmodels.DBError {
		_, err := this.newDocIfNotExist(ctx, this.orgEmailCollection, bson.M{"email": opt.Email}, body)
		return err
	}

	return withContextOfDB(f)
}

func (this *client) GetOrgEmailInfo(email string) (*dbmodels.OrgEmailCreateInfo, *dbmodels.DBError) {
	var v cOrgEmail

	f := func(ctx context.Context) *dbmodels.DBError {
		return this.getDoc1(ctx, this.orgEmailCollection, bson.M{"email": email}, bson.M{"email": 0}, &v)
	}

	if err := withContextOfDB(f); err != nil {
		return nil, err
	}

	return &dbmodels.OrgEmailCreateInfo{
		Platform: v.Platform,
		Token:    v.Token,
	}, nil
}
