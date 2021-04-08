package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (c *client) elemFilterOfIndividualSigning(email string) (bson.M, dbmodels.IDBError) {
	encryptedEmail, err := c.encrypt.encryptStr(email)
	if err != nil {
		return nil, err
	}

	return bson.M{
		fieldCorpID: genCorpID(email),
		fieldEmail:  encryptedEmail,
	}, nil
}

func docFilterOfSigning(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func (this *client) SignIndividualCLA(linkID string, info *dbmodels.IndividualSigningInfo) dbmodels.IDBError {
	email, err := this.encrypt.encryptStr(info.Email)
	if err != nil {
		return err
	}

	si, err := this.encrypt.encryptSigningInfo(&info.Info)
	if err != nil {
		return err
	}

	signing := dIndividualSigning{
		CLALanguage: info.CLALanguage,
		CorpID:      genCorpID(info.Email),
		ID:          info.ID,
		Name:        info.Name,
		Email:       email,
		Date:        info.Date,
		Enabled:     info.Enabled,
	}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}
	doc[fieldInfo] = si

	elemFilter := bson.M{
		fieldCorpID: genCorpID(info.Email),
		fieldEmail:  email,
	}
	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, false, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(ctx, this.individualSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func (this *client) DeleteIndividualSigning(linkID, email string) dbmodels.IDBError {
	elemFilter, err := this.elemFilterOfIndividualSigning(email)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pullArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			docFilterOfSigning(linkID), elemFilter,
		)
	}

	return withContext1(f)
}

func (this *client) UpdateIndividualSigning(linkID, email string, enabled bool) dbmodels.IDBError {
	elemFilter, err := this.elemFilterOfIndividualSigning(email)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateArrayElem(
			ctx, this.individualSigningCollection, fieldSignings, docFilter,
			elemFilter, bson.M{fieldEnabled: enabled},
		)
	}

	return withContext1(f)
}

func (this *client) IsIndividualSigned(linkID, email string) (bool, dbmodels.IDBError) {
	elemFilter, err := this.elemFilterOfIndividualSigning(email)
	if err != nil {
		return false, err
	}
	elemFilter[fieldEnabled] = true

	docFilter := docFilterOfSigning(linkID)

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

	err = withContext1(f)
	return signed, err
}

func (this *client) ListIndividualSigning(linkID, corpEmail, claLang string) ([]dbmodels.IndividualSigningBasicInfo, dbmodels.IDBError) {
	docFilter := docFilterOfSigning(linkID)

	arrayFilter := bson.M{}
	if corpEmail != "" {
		arrayFilter[fieldCorpID] = genCorpID(corpEmail)
	}
	if claLang != "" {
		arrayFilter[fieldLang] = claLang
	}

	project := bson.M{
		memberNameOfSignings(fieldID):      1,
		memberNameOfSignings(fieldEmail):   1,
		memberNameOfSignings(fieldName):    1,
		memberNameOfSignings(fieldEnabled): 1,
		memberNameOfSignings(fieldDate):    1,
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
		return nil, nil
	}

	docs := v[0].Signings
	r := make([]dbmodels.IndividualSigningBasicInfo, 0, len(docs))
	for i := range docs {
		item := &docs[i]

		email, err := this.encrypt.decryptStr(item.Email)
		if err != nil {
			return nil, err
		}

		r = append(r, dbmodels.IndividualSigningBasicInfo{
			ID:      item.ID,
			Email:   email,
			Name:    item.Name,
			Enabled: item.Enabled,
			Date:    item.Date,
		})
	}

	return r, nil
}
