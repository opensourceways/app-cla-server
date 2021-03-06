package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) getCorpSigning(linkID, email string) (*dCorpSigning, dbmodels.IDBError) {
	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email),
			nil, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, nil
	}

	return &signings[0], nil
}

func (this *client) DeleteCorpSigning(linkID, email string) dbmodels.IDBError {
	data, err := this.getCorpSigning(linkID, email)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}

	doc, err := structToMap(data)
	if err != nil {
		return err
	}
	doc[fieldInfo] = data.SigningInfo

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.moveArrayElem(
			ctx, this.corpSigningCollection, fieldSignings, fieldDeleted,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email), doc,
		)
	}

	return withContext1(f)
}

func (this *client) ListDeletedCorpSignings(linkID string) ([]dbmodels.CorporationSigningBasicInfo, dbmodels.IDBError) {
	key := func(k string) string {
		return fmt.Sprintf("%s.%s", fieldDeleted, k)
	}

	project := bson.M{
		key(fieldEmail): 1,
		key(fieldName):  1,
		key(fieldCorp):  1,
		key(fieldDate):  1,
		key(fieldLang):  1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldDeleted,
			docFilterOfSigning(linkID), nil, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	deleted := v[0].Deleted
	n := len(deleted)
	if n == 0 {
		return nil, nil
	}

	r := make([]dbmodels.CorporationSigningBasicInfo, 0, n)
	for i := 0; i < n; i++ {
		bi, err := this.toDBModelCorporationSigningBasicInfo(&deleted[i])
		if err != nil {
			return nil, err
		}

		r = append(r, *bi)
	}

	return r, nil
}
