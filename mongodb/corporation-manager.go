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
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCorpManager(linkID string) bson.M {
	return docFilterOfSigning(linkID)
}

func (c *client) elemFilterOfCorpManager(email string) (bson.M, dbmodels.IDBError) {
	return c.elemFilterOfIndividualSigning(email)
}

func memberNameOfCorpManager(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpManagers, field)
}

func (this *client) AddCorpAdministrator(linkID string, opt *dbmodels.CorporationManagerCreateOption) dbmodels.IDBError {
	email, err := this.encrypt.encryptStr(opt.Email)
	if err != nil {
		return err
	}

	info := dCorpManager{
		ID:       opt.ID,
		Name:     opt.Name,
		Email:    email,
		Role:     dbmodels.RoleAdmin,
		Password: opt.Password,
		CorpID:   genCorpID(opt.Email),
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(
		fieldCorpManagers, false,
		bson.M{
			fieldCorpID: genCorpID(opt.Email),
			fieldRole:   dbmodels.RoleAdmin,
		},
		docFilter,
	)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, body,
		)
	}

	return withContext1(f)
}

func (this *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (map[string]dbmodels.CorporationManagerCheckResult, dbmodels.IDBError) {
	docFilter := bson.M{
		fieldLinkStatus:   linkStatusReady,
		fieldCorpManagers: bson.M{"$type": "array"},
	}

	var elemFilter bson.M
	if opt.Email != "" {
		v, err := this.elemFilterOfCorpManager(opt.Email)
		if err != nil {
			return nil, err
		}
		elemFilter = v
	} else {
		elemFilter = bson.M{
			fieldCorpID: opt.EmailSuffix,
			fieldID:     opt.ID,
		}
	}

	project := bson.M{
		fieldLinkID:                            1,
		fieldOrgIdentity:                       1,
		fieldOrgEmail:                          1,
		fieldOrgAlias:                          1,
		memberNameOfCorpManager(fieldRole):     1,
		memberNameOfCorpManager(fieldName):     1,
		memberNameOfCorpManager(fieldEmail):    1,
		memberNameOfCorpManager(fieldPassword): 1,
		memberNameOfCorpManager(fieldChanged):  1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}
	if len(v) == 0 {
		return nil, nil
	}

	result := map[string]dbmodels.CorporationManagerCheckResult{}
	for i := range v {
		doc := &v[i]
		cm := doc.Managers
		if len(cm) == 0 {
			continue
		}

		item := &cm[0]

		email, err := this.encrypt.decryptStr(item.Email)
		if err != nil {
			return nil, err
		}

		orgRepo := dbmodels.ParseToOrgRepo(doc.OrgIdentity)
		result[doc.LinkID] = dbmodels.CorporationManagerCheckResult{
			Name:             item.Name,
			Email:            email,
			Role:             item.Role,
			Password:         item.Password,
			InitialPWChanged: item.InitialPWChanged,

			OrgInfo: dbmodels.OrgInfo{
				OrgRepo: dbmodels.OrgRepo{
					Platform: orgRepo.Platform,
					OrgID:    orgRepo.OrgID,
					RepoID:   orgRepo.RepoID,
				},
				OrgEmail: doc.OrgEmail,
				OrgAlias: doc.OrgAlias,
			},
		}

	}
	return result, nil
}

func (this *client) ResetCorporationManagerPassword(linkID, email string, opt dbmodels.CorporationManagerResetPassword) dbmodels.IDBError {
	updateCmd := bson.M{
		fieldPassword: opt.NewPassword,
		fieldChanged:  true,
	}

	elemFilter, err := this.elemFilterOfCorpManager(email)
	if err != nil {
		return err
	}
	elemFilter[fieldPassword] = opt.OldPassword

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(fieldCorpManagers, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, updateCmd)
	}

	return withContext1(f)
}

func (this *client) ListCorporationManager(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, dbmodels.IDBError) {
	elemFilter := filterOfCorpID(email)
	if role != "" {
		elemFilter["role"] = role
	}

	project := bson.M{
		memberNameOfCorpManager(fieldID):    1,
		memberNameOfCorpManager(fieldName):  1,
		memberNameOfCorpManager(fieldEmail): 1,
		memberNameOfCorpManager(fieldRole):  1,
	}

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	ms := v[0].Managers
	if ms == nil {
		return nil, nil
	}

	r := make([]dbmodels.CorporationManagerListResult, 0, len(ms))
	for i := range ms {
		item := &ms[i]

		email, err := this.encrypt.decryptStr(item.Email)
		if err != nil {
			return nil, err
		}

		r = append(r, dbmodels.CorporationManagerListResult{
			ID:    item.ID,
			Name:  item.Name,
			Email: email,
			Role:  item.Role,
		})
	}
	return r, nil
}

func (this *client) GetCorporationManager(linkID, email string) (*dbmodels.CorporationManagerCheckResult, dbmodels.IDBError) {
	elemFilter, err := this.elemFilterOfCorpManager(email)
	if err != nil {
		return nil, err
	}

	project := bson.M{
		memberNameOfCorpManager(fieldPassword): 1,
	}

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	ms := v[0].Managers
	if ms == nil {
		return nil, nil
	}

	m := v[0].Managers[0]
	return &dbmodels.CorporationManagerCheckResult{
		Password: m.Password,
	}, nil
}
