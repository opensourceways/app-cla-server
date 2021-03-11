package mongodb

import (
	"context"

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

	return &signings[1], nil
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

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.moveArrayElem(
			ctx, this.corpSigningCollection, fieldSignings, fieldDeleted,
			docFilterOfSigning(linkID), elemFilterOfCorpSigning(email), doc,
		)
	}

	return withContext1(f)
}
