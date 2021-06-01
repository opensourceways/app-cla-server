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
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

var (
	errNoDBRecord = dbError{code: dbmodels.ErrNoDBRecord, err: fmt.Errorf("no record")}
)

type dbError struct {
	code dbmodels.DBErrCode
	err  error
}

func (this dbError) Error() string {
	if this.err == nil {
		return ""
	}
	return this.err.Error()
}

func (this dbError) IsErrorOf(code dbmodels.DBErrCode) bool {
	return this.code == code
}

func (this dbError) ErrCode() dbmodels.DBErrCode {
	return this.code
}

func newDBError(code dbmodels.DBErrCode, err error) dbmodels.IDBError {
	return dbError{code: code, err: err}
}

func newSystemError(err error) dbmodels.IDBError {
	return dbError{code: dbmodels.ErrSystemError, err: err}
}
