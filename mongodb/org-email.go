package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func toDocOfOrgEmail(opt *dbmodels.OrgEmailCreateInfo) (bson.M, dbmodels.IDBError) {
	info := cOrgEmail{
		Email:    opt.Email,
		Platform: opt.Platform,
	}
	body, err := structToMap(info)
	if err != nil {
		return nil, err
	}
	body[fieldToken] = opt.Token

	return body, nil
}

func (this *client) CreateOrgEmail(opt dbmodels.OrgEmailCreateInfo) dbmodels.IDBError {
	body, err := toDocOfOrgEmail(&opt)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc(ctx, this.orgEmailCollection, bson.M{fieldEmail: opt.Email}, body)
		return err
	}

	return withContext1(f)
}

func (this *client) GetOrgEmailInfo(email string) (*dbmodels.OrgEmailCreateInfo, dbmodels.IDBError) {
	var v cOrgEmail

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(ctx, this.orgEmailCollection, bson.M{fieldEmail: email}, bson.M{fieldEmail: 0}, &v)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	return &dbmodels.OrgEmailCreateInfo{
		Email:    email,
		Platform: v.Platform,
		Token:    v.Token,
	}, nil
}
