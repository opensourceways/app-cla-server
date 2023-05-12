package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func newSigningIndex(index *dbmodels.SigningIndex) *signingIndex {
	return (*signingIndex)(index)
}

type signingIndex dbmodels.SigningIndex

func (index *signingIndex) docFilterOfSigning() bson.M {
	return bson.M{
		fieldLinkID:     index.LinkId,
		fieldLinkStatus: linkStatusReady,
	}
}

func (index *signingIndex) idFilter() bson.M {
	return bson.M{fieldID: index.SigningId}
}

func (index *signingIndex) signingIdFilter() bson.M {
	return bson.M{fieldSigningId: index.SigningId}
}

func (index *signingIndex) docFilter() bson.M {
	return bson.M{
		fieldLinkID:    index.LinkId,
		fieldSigningId: index.SigningId,
	}
}
