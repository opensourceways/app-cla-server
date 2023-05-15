package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) SignEmployeeCLA(si *dbmodels.SigningIndex, info *dbmodels.IndividualSigningInfo) dbmodels.IDBError {
	signing := dIndividualSigning{
		ID:          newObjectId(),
		Name:        info.Name,
		Email:       info.Email,
		Date:        info.Date,
		Enabled:     info.Enabled,
		CLALanguage: info.CLALanguage,
		SigningInfo: info.Info,
		CorpSID:     si.SigningId,
	}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := newSigningIndex(si).docFilterOfSigning()
	arrayFilterByElemMatch(
		fieldSignings, false, elemFilterOfIndividualSigning(info.Email), docFilter,
	)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(ctx, this.individualSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func (this *client) ListEmployeeSigning(si *dbmodels.SigningIndex, claLang string) (
	[]dbmodels.IndividualSigningBasicInfo, dbmodels.IDBError,
) {
	index := newSigningIndex(si)

	var v []cIndividualSigning
	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.individualSigningCollection,
			index.docFilterOfSigning(),
			bson.M{
				memberNameOfSignings(fieldID):      1,
				memberNameOfSignings(fieldEmail):   1,
				memberNameOfSignings(fieldName):    1,
				memberNameOfSignings(fieldEnabled): 1,
				memberNameOfSignings(fieldDate):    1,
			},
			map[string]func() bson.M{
				fieldSignings: func() bson.M {
					m := index.signingIdFilter()
					if claLang != "" {
						m[fieldLang] = claLang
					}

					return conditionTofilterArray(m)
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

func (this *client) UpdateEmployeeSigning(si *dbmodels.SigningIndex, enabled bool) (
	info dbmodels.IndividualSigningBasicInfo, err dbmodels.IDBError,
) {
	index := newSigningIndex(si)

	var v cIndividualSigning

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateAndReturnArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			index.docFilterOfSigning(), index.idFilter(),
			bson.M{fieldEnabled: enabled}, &v,
		)
	}

	if err = withContext1(f); err != nil {
		return
	}

	if len(v.Signings) == 0 {
		err = errNotFound
	} else {
		info = toIndividualSigningBasicInfo(&v.Signings[0])
	}

	return
}
