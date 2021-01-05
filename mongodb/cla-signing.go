package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) InitializeIndividualSigning(linkID string) dbmodels.IDBError {
	docFilter := bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc1(ctx, this.individualSigningCollection, docFilter, docFilter)
		return err
	}

	return withContext1(f)
}
