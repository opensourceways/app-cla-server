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
		return c.updateArrayElem(
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

	claOrg, err := getSigningDoc(v, func(doc *CLAOrg) bool {
		return len(doc.Corporations) > 0
	})

	item := claOrg.Corporations[0]
	if !item.PDFUploaded {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrPDFHasNotUploaded,
			Err:     fmt.Errorf("pdf has not yet been uploaded"),
		}
	}

	return item.PDF, nil
}
