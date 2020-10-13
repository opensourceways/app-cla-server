package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func (c *client) UploadCorporationSigningPDF(claOrgID, adminEmail string, pdf []byte) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForCorpSigning(filter)

		update := bson.M{"$set": bson.M{
			fmt.Sprintf("%s.$[elem].pdf", fieldCorporations):          pdf,
			fmt.Sprintf("%s.$[elem].pdf_uploaded", fieldCorporations): true,
		}}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"elem.corp_id": util.EmailSuffix(adminEmail),
					},
				},
			},
		}

		r, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return dbmodels.DBError{
				ErrCode: util.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		if r.ModifiedCount == 0 {
			return dbmodels.DBError{
				ErrCode: util.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the corp signing record or upload repeatedly"),
			}
		}
		return nil
	}

	return withContext(f)
}

func (c *client) DownloadCorporationSigningPDF(claOrgID, email string) ([]byte, error) {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return nil, err
	}

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForCorpSigning(filter)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldCorporations: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorporations),
					"cond":  bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
				}},
			}},
			bson.M{"$project": bson.M{
				corpSigningField("pdf"):          1,
				corpSigningField("pdf_uploaded"): 1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}

	cs := v[0].Corporations
	if len(cs) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrInvalidParameter,
			Err:     fmt.Errorf("can't find the corp signing in this record"),
		}
	}

	item := cs[0]
	if !item.PDFUploaded {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrPDFHasNotUploaded,
			Err:     fmt.Errorf("pdf has not yet been uploaded"),
		}
	}

	return item.PDF, nil
}
