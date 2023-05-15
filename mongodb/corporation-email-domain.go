package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) AddCorpEmailDomain(si *dbmodels.SigningIndex, domain string) dbmodels.IDBError {
	index := newSigningIndex(si)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushNestedArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			index.docFilterOfSigning(),
			index.idFilter(), bson.M{fieldDomains: domain},
		)
	}

	return withContext1(f)
}

func (this *client) GetCorpEmailDomains(si *dbmodels.SigningIndex) ([]string, dbmodels.IDBError) {
	index := newSigningIndex(si)

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.corpSigningCollection, index.docFilterOfSigning(),
			bson.M{
				memberNameOfSignings(fieldDomains): 1,
			},
			map[string]func() bson.M{
				fieldSignings: func() bson.M {
					return conditionTofilterArray(index.idFilter())
				},
			},
			&v,
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

	return signings[0].Domains, nil
}
