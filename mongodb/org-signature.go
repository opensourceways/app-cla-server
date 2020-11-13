package mongodb

import (
	"context"
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (this *client) UploadOrgSignature(orgCLAID string, pdf []byte) error {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return this.updateDoc(
			ctx, this.orgCLACollection, filterOfDocID(oid),
			bson.M{
				fieldOrgSignature:    pdf,
				fieldOrgSignatureTag: util.Md5sumOfBytes(pdf),
			},
		)
	}

	return withContext(f)
}

func (this *client) DownloadOrgSignature(orgCLAID string) ([]byte, error) {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return nil, err
	}

	var v OrgCLA

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.orgCLACollection, filterOfDocID(oid),
			bson.M{fieldOrgSignature: 1},
			&v,
		)
	}

	if withContext(f); err != nil {
		return nil, err
	}

	return v.OrgSignature, nil
}

func (this *client) DownloadOrgSignatureByMd5(orgCLAID, md5sum string) ([]byte, error) {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return nil, err
	}

	var v []OrgCLA

	f := func(ctx context.Context) error {
		pipeline := bson.A{
			bson.M{"$match": filterOfDocID(oid)},
			bson.M{"$project": bson.M{
				fieldOrgSignature: bson.M{"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{md5sum, "$" + fieldOrgSignatureTag}},
					"then": "$$REMOVE",
					"else": "$" + fieldOrgSignature,
				}},
			}},
		}

		col := this.collection(this.orgCLACollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	if withContext(f); err != nil {
		return nil, err
	}

	if len(v) > 0 {
		return v[0].OrgSignature, nil
	}

	return nil, dbmodels.DBError{
		Err:     fmt.Errorf("can't find org's cla"),
		ErrCode: util.ErrNoDBRecord,
	}
}
