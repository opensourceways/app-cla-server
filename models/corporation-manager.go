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

package models

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerAuthentication struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func (this CorporationManagerAuthentication) Authenticate() (map[string]dbmodels.CorporationManagerCheckResult, IModelError) {
	info := dbmodels.CorporationManagerCheckInfo{}
	if merr := checkEmailFormat(this.User); merr == nil {
		info.Email = this.User
	} else {
		if merr := checkManagerID(this.User); merr != nil {
			return nil, merr
		}

		i := strings.LastIndex(this.User, "_")
		info.EmailSuffix = this.User[(i + 1):]
		info.ID = this.User[:i]
	}

	v, err := dbmodels.GetDB().CheckCorporationManagerExist(info)
	if err == nil {
		for k := range v {
			if !isSamePasswords(v[k].Password, this.Password) {
				delete(v, k)
			}
		}
		return v, nil
	}

	return nil, parseDBError(err)
}

func CreateCorporationAdministrator(linkID, name, email string) (*dbmodels.CorporationManagerCreateOption, IModelError) {
	pw := newPWForCorpManager()
	encryptedPW, merr := encryptPassword(pw)
	if merr != nil {
		return nil, merr
	}

	opt := &dbmodels.CorporationManagerCreateOption{
		ID:       "admin",
		Name:     name,
		Email:    email,
		Password: encryptedPW,
		Role:     dbmodels.RoleAdmin,
	}
	err := dbmodels.GetDB().AddCorpAdministrator(linkID, opt)
	if err == nil {
		opt.ID = fmt.Sprintf("admin_%s", util.EmailSuffix(email))
		opt.Password = pw
		return opt, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLinkOrManagerExists, err)
	}

	return nil, parseDBError(err)
}

type CorporationManagerResetPassword dbmodels.CorporationManagerResetPassword

func (this CorporationManagerResetPassword) Validate() IModelError {
	if this.NewPassword == this.OldPassword {
		return newModelError(ErrSamePassword, fmt.Errorf("the new password is same as old one"))
	}

	n := len(this.NewPassword)
	cfg := config.AppConfig
	if n < cfg.MinLengthOfPassword || n > cfg.MaxLengthOfPassword {
		return newModelError(
			ErrTooShortOrLongPassword,
			fmt.Errorf(
				"the length of password should be between %d and %d",
				cfg.MinLengthOfPassword, cfg.MaxLengthOfPassword,
			))
	}

	return checkPassword(this.NewPassword)
}

func (this CorporationManagerResetPassword) Reset(linkID, email string) IModelError {
	pw, merr := encryptPassword(this.NewPassword)
	if merr != nil {
		return merr
	}

	record, merr := this.getCorporationManager(linkID, email)
	if merr != nil {
		return merr
	}
	if record == nil {
		return newModelError(ErrCorpManagerDoesNotExist, fmt.Errorf("corp manager does not exist"))
	}

	if !isSamePasswords(record.Password, this.OldPassword) {
		return newModelError(ErrWrongOldPassword, fmt.Errorf("old password is not correct"))
	}

	err := dbmodels.GetDB().ResetCorporationManagerPassword(
		linkID, email, dbmodels.CorporationManagerResetPassword{
			OldPassword: record.Password, NewPassword: pw,
		},
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrNoManagerOrFO, err)
	}
	return parseDBError(err)
}

func (this CorporationManagerResetPassword) getCorporationManager(linkID, email string) (*dbmodels.CorporationManagerCheckResult, IModelError) {
	v, err := dbmodels.GetDB().GetCorporationManager(linkID, email)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}

func ListCorporationManagers(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, IModelError) {
	v, err := dbmodels.GetDB().ListCorporationManager(linkID, email, role)
	if err == nil {
		if v == nil {
			v = []dbmodels.CorporationManagerListResult{}
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}

func newPWForCorpManager() string {
	return util.RandStr(8, "alphanum")
}
