package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCLAPDF(key dbmodels.CLAPDFIndex) (bson.M, dbmodels.IDBError) {
	info := dCLAPDF{
		LinkID: key.LinkID,
		Apply:  key.Apply,
		Lang:   key.Lang,
		Hash:   key.Hash,
	}
	return structToMap(info)
}

func (this *client) UploadCLAPDF(key dbmodels.CLAPDFIndex, pdf []byte) dbmodels.IDBError {
	docFilter, err := docFilterOfCLAPDF(key)
	if err != nil {
		return err
	}

	doc := bson.M{fieldPDF: pdf}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc(ctx, this.claPDFCollection, docFilter, doc)
		return err
	}

	return withContext1(f)
}

func (this *client) DownloadCLAPDF(key dbmodels.CLAPDFIndex) ([]byte, dbmodels.IDBError) {
	docFilter, err := docFilterOfCLAPDF(key)
	if err != nil {
		return nil, err
	}

	var v dCLAPDF

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc(
			ctx, this.claPDFCollection, docFilter,
			bson.M{fieldPDF: 1}, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	return v.PDF, nil
}
