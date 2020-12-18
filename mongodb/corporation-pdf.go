package mongodb

import (
	"context"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"go.mongodb.org/mongo-driver/bson"
)

func docFilterOfCorpSigningPDF(linkID string, email string) bson.M {
	return bson.M{
		fieldLinkID: linkID,
		fieldCorpID: genCorpID(email),
	}
}

func (this *client) UploadCorporationSigningPDF(linkID string, adminEmail string, pdf *[]byte) *dbmodels.DBError {
	docFilter := docFilterOfCorpSigningPDF(linkID, adminEmail)

	doc := bson.M{"pdf": *pdf}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) error {
		_, err := this.replaceDoc(ctx, this.corpPDFCollection, docFilter, doc)
		return err
	}

	if err := withContext(f); err != nil {
		return systemError(err)
	}
	return nil
}

func (this *client) DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, error) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(linkID, email), bson.M{"pdf": 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return &v.PDF, nil
}

func (this *client) IsCorpSigningPDFUploaded(linkID string, email string) (bool, error) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(linkID, email), bson.M{"_id": 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		if isErrOfNoDocument(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (this *client) ListCorpsWithPDFUploaded(linkID string) ([]string, error) {
	var v []struct {
		CorpID string `bson:"corp_id"`
	}

	f := func(ctx context.Context) error {
		return this.getDocs(
			ctx, this.corpPDFCollection,
			bson.M{fieldLinkID: linkID},
			bson.M{fieldCorpID: 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(v))
	for i := range v {
		result = append(result, v[i].CorpID)
	}
	return result, nil
}
