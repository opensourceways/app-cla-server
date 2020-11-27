package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfOrgEmail(email string) bson.M {
	return bson.M{fieldEmail: email}
}

func (this *client) CreateOrgEmail(opt *dbmodels.OrgEmailCreateInfo) error {
	info := dOrgEmail{
		Email:    opt.Email,
		Platform: opt.Platform,
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}
	body[fieldToken] = opt.Token

	f := func(ctx context.Context) error {
		_, err := this.newDocIfNotExist(ctx, this.orgEmailCollection, docFilterOfOrgEmail(opt.Email), body)
		if err != nil && isErrorOfRecordExists(err) {
			return nil
		}
		return err
	}

	return withContext(f)
}

func (this *client) GetOrgEmailInfo(email string) (*dbmodels.OrgEmailCreateInfo, error) {
	var v dOrgEmail

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.orgEmailCollection,
			docFilterOfOrgEmail(email), bson.M{fieldEmail: 0}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return &dbmodels.OrgEmailCreateInfo{
		Platform: v.Platform,
		Token:    v.Token,
	}, nil
}
