package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfIndividualSigning(email string) bson.M {
	return bson.M{
		fieldEmail: email,
	}
}

func docFilterOfSigning(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func (this *client) SignIndividualCLA(linkID string, info *dbmodels.IndividualSigningInfo) dbmodels.IDBError {
	signing := dIndividualSigning{
		ID:          newObjectId(),
		Name:        info.Name,
		Email:       info.Email,
		Date:        info.Date,
		Enabled:     info.Enabled,
		CLALanguage: info.CLALanguage,
		SigningInfo: info.Info,
	}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(
		fieldSignings, false, elemFilterOfIndividualSigning(info.Email), docFilter,
	)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(ctx, this.individualSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func (this *client) DeleteIndividualSigning(si *dbmodels.SigningIndex) (
	info dbmodels.IndividualSigningBasicInfo, err dbmodels.IDBError,
) {
	index := newSigningIndex(si)

	var v cIndividualSigning

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pullAndReturnArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			index.docFilterOfSigning(),
			index.idFilter(), &v,
		)
	}

	if err = withContext1(f); err != nil {
		return
	}

	if len(v.Signings) > 0 {
		info = toIndividualSigningBasicInfo(&v.Signings[0])
	}

	return
}

func (this *client) IsIndividualSigned(linkID, email string) (bool, dbmodels.IDBError) {
	docFilter := docFilterOfSigning(linkID)

	elemFilter := elemFilterOfIndividualSigning(email)
	elemFilter[fieldEnabled] = true

	signed := false
	f := func(ctx context.Context) dbmodels.IDBError {
		v, err := this.isArrayElemNotExists(
			ctx, this.individualSigningCollection, fieldSignings, docFilter, elemFilter,
		)
		if err != nil {
			return newSystemError(err)
		}
		signed = !v
		return nil
	}

	err := withContext1(f)
	return signed, err
}

func (this *client) ListIndividualSigning(linkID, claLang string) (
	[]dbmodels.IndividualSigningBasicInfo, dbmodels.IDBError,
) {
	var v []cIndividualSigning
	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.individualSigningCollection,
			docFilterOfSigning(linkID),
			bson.M{
				memberNameOfSignings(fieldID):      1,
				memberNameOfSignings(fieldEmail):   1,
				memberNameOfSignings(fieldName):    1,
				memberNameOfSignings(fieldEnabled): 1,
				memberNameOfSignings(fieldDate):    1,
			},
			map[string]func() bson.M{
				fieldSignings: func() bson.M {
					if claLang != "" {
						return conditionTofilterArray(bson.M{fieldLang: claLang})
					}

					return bson.M{"$toBool": 1}
				},
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, nil
	}

	docs := v[0].Signings
	r := make([]dbmodels.IndividualSigningBasicInfo, len(docs))
	for i := range docs {
		r[i] = toIndividualSigningBasicInfo(&docs[i])
	}

	return r, nil
}

func toIndividualSigningBasicInfo(doc *dIndividualSigning) dbmodels.IndividualSigningBasicInfo {
	return dbmodels.IndividualSigningBasicInfo{
		ID:      doc.ID,
		Email:   doc.Email,
		Name:    doc.Name,
		Enabled: doc.Enabled,
		Date:    doc.Date,
	}
}
