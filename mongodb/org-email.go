/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	t, err := this.encrypt.encryptBytes(opt.Token)
	if err != nil {
		return err
	}
	body[fieldToken] = t

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

func (this *client) GetOrgEmailOfLink(linkID string) (*dbmodels.OrgEmailCreateInfo, dbmodels.IDBError) {
	var v cLink
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.linkCollection,
			bson.M{
				fieldLinkID:     linkID,
				fieldLinkStatus: linkStatusReady,
			},
			bson.M{fieldOrgEmail: 1}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	oe := &v.OrgEmail

	t, err := this.encrypt.decryptBytes(oe.Token)
	if err != nil {
		return nil, err
	}

	return &dbmodels.OrgEmailCreateInfo{
		Email:    oe.Email,
		Platform: oe.Platform,
		Token:    t,
	}, nil
}
