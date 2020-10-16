package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func (c *client) UploadCorporationSigningPDF(claOrgID, adminEmail string, pdf []byte) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return c.updateArryItem(
			ctx, claOrgCollection, fieldCorporations,
			filterOfDocID(oid),
			filterOfCorpID(adminEmail),
			bson.M{
				"pdf":          pdf,
				"pdf_uploaded": true,
			}, true,
		)
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
		return c.getArrayElem(
			ctx, claOrgCollection, fieldCorporations,
			filterOfDocID(oid),
			filterOfCorpID(email),
			bson.M{
				corpSigningField("pdf"):          1,
				corpSigningField("pdf_uploaded"): 1,
			}, &v,
		)
	}

	if err = withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNoDBRecord,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}

	cs := v[0].Corporations
	if len(cs) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNoDBRecord,
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
