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

package obs

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func NewFileStorage(c OBS) dbmodels.IFile {
	return fileStorage{c: c}
}

type fileStorage struct {
	c OBS
}

func (fs fileStorage) UploadCorporationSigningPDF(linkID, adminEmail string, pdf []byte) dbmodels.IDBError {
	err := fs.c.WriteObject(buildCorpSigningPDFPath(linkID, adminEmail), pdf)
	return toDBError(err)
}

func (fs fileStorage) DownloadCorporationSigningPDF(linkID, email, path string) dbmodels.IDBError {
	err := fs.c.ReadObject(buildCorpSigningPDFPath(linkID, email), path)
	if err == nil {
		return nil
	}

	if err.IsObjectNotFound() {
		return dbmodels.NewDBError(dbmodels.ErrNoDBRecord, err)
	}
	return toDBError(err)
}

func (fs fileStorage) IsCorporationSigningPDFUploaded(linkID, email string) (bool, dbmodels.IDBError) {
	b, err := fs.c.HasObject(buildCorpSigningPDFPath(linkID, email))
	return b, toDBError(err)
}

func (fs fileStorage) ListCorporationsWithPDFUploaded(linkID string) ([]string, dbmodels.IDBError) {
	prefix := buildCorpSigningPDFPath(linkID, "")

	r, err := fs.c.ListObject(prefix)
	if err != nil {
		return nil, toDBError(err)
	}

	result := make([]string, 0, len(r))
	for _, item := range r {
		result = append(result, strings.TrimPrefix(item, prefix))
	}
	return result, nil
}

func buildCorpSigningPDFPath(linkID string, email string) string {
	return fmt.Sprintf("%s/%s", linkID, util.EmailSuffix(email))
}

func toDBError(err error) dbmodels.IDBError {
	if err == nil {
		return nil
	}
	return dbmodels.NewDBError(dbmodels.ErrSystemError, err)
}
