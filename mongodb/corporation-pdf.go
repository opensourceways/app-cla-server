package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type dCorpSigningPDF struct {
	OrgIdentity string `bson:"org_identity" json:"org_identity" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF         []byte `bson:"pdf" json:"pdf,omitempty"`
}

func docFilterOfCorpSigningPDF(orgRepo *dbmodels.OrgRepo, email string) bson.M {
	return bson.M{
		"org_identity":     orgIdentity(orgRepo),
		fieldCorporationID: genCorpID(email),
	}
}

func (this *client) UploadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, adminEmail string, pdf *[]byte) error {
	docFilter := docFilterOfCorpSigningPDF(orgRepo, adminEmail)

	doc := bson.M{"pdf": *pdf}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) error {
		_, err := this.replaceDoc(ctx, this.corpPDFCollection, docFilter, doc)
		return err
	}

	return withContext(f)
}

func (this *client) DownloadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, email string) (*[]byte, error) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(orgRepo, email), bson.M{"pdf": 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return &v.PDF, nil
}

func (this *client) IsCorpSigningPDFUploaded(orgRepo *dbmodels.OrgRepo, email string) (bool, error) {
	var v dCorpSigningPDF

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.corpPDFCollection,
			docFilterOfCorpSigningPDF(orgRepo, email), bson.M{"_id": 1}, &v,
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

func (this *client) ListCorpsWithPDFUploaded(orgRepo *dbmodels.OrgRepo) ([]string, error) {
	var v []struct {
		CorpID string `bson:"corp_id"`
	}

	f := func(ctx context.Context) error {
		return this.getDocs(
			ctx, this.corpPDFCollection,
			bson.M{"org_identity": orgIdentity(orgRepo)},
			bson.M{fieldCorporationID: 1}, &v,
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
