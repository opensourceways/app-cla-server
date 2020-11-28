package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (c *client) UploadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, email string, pdf []byte) error {
	elemFilter := elemFilterOfCorpSigning(email)

	docFilter := docFilterOfCorpSigning(orgRepo)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) error {
		return c.updateArrayElem(
			ctx, c.corpSigningCollection, fieldSignings, docFilter, elemFilter,
			bson.M{
				"pdf":          pdf,
				"pdf_uploaded": true,
			}, false,
		)
	}

	return withContext(f)

}

func (c *client) DownloadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, email string) ([]byte, error) {
	elemFilter := elemFilterOfCorpSigning(email)

	docFilter := docFilterOfCorpSigning(orgRepo)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, c.corpSigningCollection, fieldSignings, docFilter, elemFilter,
			bson.M{
				memberNameOfSignings("pdf"):          1,
				memberNameOfSignings("pdf_uploaded"): 1,
			}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	}

	item := v[0].Signings[0]
	if !item.PDFUploaded {
		return nil, nil
	}
	return item.PDF, nil
}
