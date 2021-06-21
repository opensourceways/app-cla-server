package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) AddCorpEmailDomain(linkID, adminEmail, subEmail string) dbmodels.IDBError {
	elemFilter := elemFilterOfCorpSigning(adminEmail)

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushNestedArrayElem(
			ctx, this.corpSigningCollection, fieldSignings, docFilter,
			elemFilter, bson.M{fieldDomains: genCorpID(subEmail)},
		)
	}

	return withContext1(f)
}

func (this *client) GetCorpSigningEmailDomains(linkID, email string) ([]string, dbmodels.IDBError) {
	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem1(
			ctx, this.corpSigningCollection, fieldSignings,
			docFilterOfSigning(linkID),
			bson.M{
				memberNameOfSignings(fieldDomains): 1,
			},
			func() bson.M {
				return bson.M{"$and": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$$this.%s", fieldDomains)},
					bson.M{"$in": bson.A{genCorpID(email), fmt.Sprintf("$$this.%s", fieldDomains)}},
				}}
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
