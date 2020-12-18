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

func (this *client) SignAsIndividual(linkID string, info *dbmodels.IndividualSigningInfo) *dbmodels.DBError {
	signing := dIndividualSigning{
		CLALanguage: info.CLALanguage,
		CorpID:      genCorpID(info.Email),
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

	f := func(ctx context.Context) *dbmodels.DBError {
		return this.pushArrayElem(ctx, this.individualSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContextOfDB(f)
}

func (this *client) DeleteIndividualSigning(linkID, email string) error {
	f := func(ctx context.Context) error {
		return this.pullArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			docFilterOfSigning(linkID),
			elemFilterOfIndividualSigning(email),
		)
	}

	return withContext(f)
}

func (this *client) UpdateIndividualSigning(linkID, email string, enabled bool) error {
	elemFilter := elemFilterOfIndividualSigning(email)

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) error {
		return this.updateArrayElem(
			ctx, this.individualSigningCollection, fieldSignings, docFilter,
			elemFilter, bson.M{"enabled": enabled},
		)
	}

	return withContext(f)
}

func (this *client) IsIndividualSigned(orgRepo *dbmodels.OrgRepo, email string) (bool, *dbmodels.DBError) {
	if orgRepo.RepoID == "" {
		return this.isIndividualSignedToOrg(orgRepo, email)
	}

	identity := orgRepo.String()

	docFilter := bson.M{
		fieldLinkStatus: linkStatusReady,
		fieldOrgIdentity: bson.M{"$in": bson.A{
			dbmodels.OrgRepo{Platform: orgRepo.Platform, OrgID: orgRepo.OrgID}.String(),
			identity,
		}},
	}

	fieldEnabled := memberNameOfSignings("enabled")

	elemFilter := elemFilterOfIndividualSigning(email)
	elemFilter[fieldEnabled] = true

	var v []cIndividualSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.individualSigningCollection, fieldSignings, docFilter, elemFilter,
			bson.M{
				fieldOrgIdentity: 1,
				fieldEnabled:     1,
			}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return false, systemError(err)
	}

	num := len(v)
	if num == 0 {
		return false, nil
	}
	if num == 1 {
		return len(v[0].Signings) > 0, nil
	}

	for i := 0; i < len(v); i++ {
		if v[i].OrgIdentity == identity {
			return len(v[i].Signings) > 0, nil
		}
	}

	return false, nil
}

func (this *client) isIndividualSignedToOrg(orgRepo *dbmodels.OrgRepo, email string) (bool, *dbmodels.DBError) {
	docFilter := bson.M{
		fieldLinkStatus:  linkStatusReady,
		fieldOrgIdentity: orgRepo.String(),
	}

	elemFilter := elemFilterOfIndividualSigning(email)
	elemFilter[memberNameOfSignings("enabled")] = true

	signed := false
	f := func(ctx context.Context) *dbmodels.DBError {
		v, err := this.isArrayElemNotExists(
			ctx, this.individualSigningCollection, fieldSignings, docFilter, elemFilter,
		)
		signed = v
		return err
	}

	err := withContextOfDB(f)
	return signed, err
}

func (this *client) ListIndividualSigning(linkID, corpEmail, claLang string) ([]dbmodels.IndividualSigningBasicInfo, error) {
	docFilter := docFilterOfSigning(linkID)

	arrayFilter := bson.M{fieldCorpID: genCorpID(corpEmail)}
	if claLang != "" {
		arrayFilter[fieldCLALang] = claLang
	}

	project := bson.M{
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
		return nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, nil
	}

	docs := v[0].Signings
	r := make([]dbmodels.IndividualSigningBasicInfo, 0, len(docs))
	for _, item := range docs {
		r = append(r, dbmodels.IndividualSigningBasicInfo{
			Email:   item.Email,
			Name:    item.Name,
			Enabled: item.Enabled,
			Date:    item.Date,
		})
	}

	return r, nil
}
