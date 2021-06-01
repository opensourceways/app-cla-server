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

func docFilterOfCorpSigningPDF(linkID string, email string) bson.M {
	return bson.M{
		fieldLinkID: linkID,
		fieldCorpID: genCorpID(email),
	}
}

func (this *client) UploadCorporationSigningPDF(linkID string, adminEmail string, pdf *[]byte) dbmodels.IDBError {
	docFilter := docFilterOfCorpSigningPDF(linkID, adminEmail)

	doc := bson.M{fieldPDF: *pdf}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc(ctx, this.corpPDFCollection, docFilter, doc)
		return err
	}

	return withContext1(f)
}

func (this *client) DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, dbmodels.IDBError) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(linkID, email), bson.M{fieldPDF: 1}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	return &v.PDF, nil
}

func (this *client) IsCorpSigningPDFUploaded(linkID string, email string) (bool, dbmodels.IDBError) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(linkID, email), bson.M{"_id": 1}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (this *client) ListCorpsWithPDFUploaded(linkID string) ([]string, dbmodels.IDBError) {
	var v []struct {
		CorpID string `bson:"corp_id"`
	}

	f := func(ctx context.Context) error {
		return this.getDocs(
			ctx, this.corpPDFCollection,
			bson.M{fieldLinkID: linkID},
			bson.M{fieldCorpID: 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	result := make([]string, 0, len(v))
	for i := range v {
		result = append(result, v[i].CorpID)
	}
	return result, nil
}
