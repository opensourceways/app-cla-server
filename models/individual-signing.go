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

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func InitializeIndividualSigning(linkID string, cla *CLAInfo) IModelError {
	err := dbmodels.GetDB().InitializeIndividualSigning(linkID, cla)
	return parseDBError(err)
}

type IndividualSigning dbmodels.IndividualSigningInfo

func (this *IndividualSigning) Validate(userID, email string) IModelError {
	if this.Email != email {
		return newModelError(ErrUnmatchedEmail, fmt.Errorf("unmatched email"))
	}

	if this.ID != userID {
		return newModelError(ErrUnmatchedUserID, fmt.Errorf("unmatched user id"))
	}

	return nil
}

func (this *IndividualSigning) Create(linkID string, enabled bool) IModelError {
	this.Date = util.Date()
	this.Enabled = enabled

	err := dbmodels.GetDB().SignIndividualCLA(
		linkID, (*dbmodels.IndividualSigningInfo)(this),
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResigned, err)
	}
	return parseDBError(err)
}

func IsIndividualSigned(linkID, email string) (bool, IModelError) {
	b, err := dbmodels.GetDB().IsIndividualSigned(linkID, email)
	if err == nil {
		return b, nil
	}
	return b, parseDBError(err)
}
