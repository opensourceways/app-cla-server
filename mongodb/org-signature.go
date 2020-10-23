package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (c *client) UploadOrgSignature(orgCLAID string, pdf []byte) error {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return c.updateDoc(
			ctx, orgCLACollection, filterOfDocID(oid),
			bson.M{
				fieldOrgSignature:    pdf,
				fieldOrgSignatureTag: true,
			},
		)
	}

	return withContext(f)
}

func (c *client) DownloadOrgSignature(orgCLAID string) ([]byte, error) {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return nil, err
	}

	var v OrgCLA

	f := func(ctx context.Context) error {
		return c.getDoc(
			ctx, orgCLACollection, filterOfDocID(oid),
			bson.M{
				fieldOrgSignature:    1,
				fieldOrgSignatureTag: 1,
			}, &v,
		)
	}

	if withContext(f); err != nil {
		return nil, err
	}

	return v.OrgSignature, nil
}
