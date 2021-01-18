package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfIndividualSigning(email string) bson.M {
	return bson.M{
		fieldCorpID: genCorpID(email),
		"email":     email,
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
		CLALanguage: info.CLALanguage,
		CorpID:      genCorpID(info.Email),
		ID:          info.ID,
		Name:        info.Name,
		Email:       info.Email,
		Date:        info.Date,
		Enabled:     info.Enabled,
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

func (this *client) DeleteIndividualSigning(linkID, email string) dbmodels.IDBError {
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pullArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			docFilterOfSigning(linkID),
			elemFilterOfIndividualSigning(email),
		)
	}

	return withContext1(f)
}

func (this *client) UpdateIndividualSigning(linkID, email string, enabled bool) dbmodels.IDBError {
	elemFilter := elemFilterOfIndividualSigning(email)

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateArrayElem(
			ctx, this.individualSigningCollection, fieldSignings, docFilter,
			elemFilter, bson.M{"enabled": enabled},
		)
	}

	return withContext1(f)
}

func (this *client) IsIndividualSigned(linkID, email string) (bool, dbmodels.IDBError) {
	docFilter := docFilterOfSigning(linkID)

	elemFilter := elemFilterOfIndividualSigning(email)
	elemFilter[memberNameOfSignings("enabled")] = true

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

func (this *client) ListIndividualSigning(linkID, corpEmail, claLang string) ([]dbmodels.IndividualSigningBasicInfo, dbmodels.IDBError) {
	docFilter := docFilterOfSigning(linkID)

	arrayFilter := bson.M{fieldCorpID: genCorpID(corpEmail)}
	if claLang != "" {
		arrayFilter[fieldCLALang] = claLang
	}

	project := bson.M{
		memberNameOfSignings("id"):      1,
		memberNameOfSignings("email"):   1,
		memberNameOfSignings("name"):    1,
		memberNameOfSignings("enabled"): 1,
		memberNameOfSignings("date"):    1,
	}

	var v []cIndividualSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			docFilter, arrayFilter, project, &v)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord1
	}

	docs := v[0].Signings
	r := make([]dbmodels.IndividualSigningBasicInfo, 0, len(docs))
	for i := range docs {
		item := &docs[i]
		r = append(r, dbmodels.IndividualSigningBasicInfo{
			ID:      item.ID,
			Email:   item.Email,
			Name:    item.Name,
			Enabled: item.Enabled,
			Date:    item.Date,
		})
	}

	return r, nil
}
