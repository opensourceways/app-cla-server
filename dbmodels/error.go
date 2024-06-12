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

package dbmodels

type DBErrCode string

type IDBError interface {
	Error() string
	IsErrorOf(DBErrCode) bool
	ErrCode() DBErrCode
}

const (
	ErrSystemError       DBErrCode = "system_error"
	ErrNoDBRecord        DBErrCode = "no_db_record"
	ErrRecordExists      DBErrCode = "db_record_exists"
	ErrMarshalDataFaield DBErrCode = "failed_to_marshal_data"
)

type dbError struct {
	code DBErrCode
	err  error
}

func (e dbError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e dbError) IsErrorOf(code DBErrCode) bool {
	return e.code == code
}

func (e dbError) ErrCode() DBErrCode {
	return e.code
}

func NewDBError(code DBErrCode, err error) IDBError {
	return dbError{code: code, err: err}
}
