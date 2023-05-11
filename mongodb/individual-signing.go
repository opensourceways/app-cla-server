package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfIndividualSigning(email string) bson.M {
	return bson.M{
		fieldCorpID: genCorpID(email),
		fieldEmail:  email,
	}
}

func docFilterOfSigning(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func newSigningIndex(index *dbmodels.SigningIndex) *signingIndex {
	return (*signingIndex)(index)
}

type signingIndex dbmodels.SigningIndex

func (index *signingIndex) docFilterOfSigning() bson.M {
	return bson.M{
		fieldLinkID:     index.LinkId,
		fieldLinkStatus: linkStatusReady,
	}
}

func (index *signingIndex) signingItemFilter() bson.M {
	return bson.M{fieldID: index.SigningId}
}

func (index *signingIndex) signingIdFilter() bson.M {
	return bson.M{fieldSigningId: index.SigningId}
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
			elemFilter, bson.M{fieldEnabled: enabled},
		)
	}

	return withContext1(f)
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

func (this *client) ListIndividualSigning(linkID, corpEmail, claLang string) ([]dbmodels.IndividualSigningBasicInfo, dbmodels.IDBError) {
	docFilter := docFilterOfSigning(linkID)

	var domains []string
	if corpEmail != "" {
		v, err := this.GetCorpEmailDomains(linkID, corpEmail)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, nil
		}
		domains = v
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
		return this.getArrayElems(
			ctx, this.individualSigningCollection, docFilter, project,
			map[string]func() bson.M{
				fieldSignings: func() bson.M {
					cond := bson.A{}
					if claLang != "" {
						cond = append(cond, bson.M{"$eq": bson.A{"$$this." + fieldLang, claLang}})
					}
					if len(domains) > 0 {
						cond = append(cond, bson.M{"$in": bson.A{fmt.Sprintf("$$this.%s", fieldCorpID), domains}})
					}

					n := len(cond)
					if n > 1 {
						return bson.M{"$and": cond}
					}
					if n > 0 {
						return cond[0].(bson.M)
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
