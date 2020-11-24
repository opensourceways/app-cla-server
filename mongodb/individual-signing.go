package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfEnabledLink(platform, org, repo string) bson.M {
	return bson.M{
		fieldOrgIdentity: orgIdentity(platform, org, repo),
		fieldLinkStatus:  linkStatusEnabled,
	}
}

func docFilterOfIndividualSigning(platform, org, repo string) bson.M {
	return docFilterOfEnabledLink(platform, org, repo)
}

func elemFilterOfIndividualSigning(email string) bson.M {
	return bson.M{
		fieldCorpID: genCorpID(email),
		"email":     email,
	}
}

func (this *client) SignAsIndividual(platform, org, repo string, info dbmodels.IndividualSigningInfo) error {
	signing := dIndividualSigning{
		CLALanguage: "",
		CorpID:      genCorpID(info.Email),
		Name:        info.Name,
		Email:       info.Email,
		Date:        info.Date,
		Enabled:     info.Enabled,
		SigningInfo: info.Info,
	}
	body, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfIndividualSigning(platform, org, repo)
	arrayFilterByElemMatch(
		fieldSignings, false, elemFilterOfIndividualSigning(info.Email), docFilter,
	)

	f := func(ctx context.Context) error {
		return this.pushArrayElem(ctx, this.individualSigningCollection, fieldSignings, docFilter, body)
	}

	return withContext(f)
}

func (this *client) DeleteIndividualSigning(platform, org, repo, email string) error {
	f := func(ctx context.Context) error {
		return this.pullArrayElem(
			ctx, this.individualSigningCollection, fieldSignings,
			docFilterOfIndividualSigning(platform, org, repo),
			elemFilterOfIndividualSigning(email),
		)
	}

	return withContext(f)
}

func (this *client) UpdateIndividualSigning(platform, org, repo, email string, enabled bool) error {
	elemFilter := elemFilterOfIndividualSigning(email)
	docFilter := docFilterOfIndividualSigning(platform, org, repo)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) error {
		return this.updateArrayElem(
			ctx, this.individualSigningCollection, fieldSignings, docFilter,
			elemFilter, bson.M{"enabled": enabled}, false,
		)
	}

	return withContext(f)
}

func (this *client) isIndividualSignedToOrg(platform, org, email string) (bool, error) {
	docFilter := docFilterOfIndividualSigning(platform, org, "")

	elemFilter := elemFilterOfIndividualSigning(email)
	elemFilter[memberNameOfSignings("enabled")] = true

	signed := false
	f := func(ctx context.Context) error {
		v, err := this.isArrayElemNotExists(
			ctx, this.individualSigningCollection, fieldSignings, docFilter, elemFilter,
		)
		signed = v
		return err
	}

	err := withContext(f)
	return signed, err
}

func (this *client) IsIndividualSigned(platform, org, repo, email string) (bool, error) {
	if repo == "" {
		return this.isIndividualSignedToOrg(platform, org, email)
	}

	identity := orgIdentity(platform, org, repo)

	docFilter := bson.M{
		fieldLinkStatus:  linkStatusEnabled,
		fieldOrgIdentity: bson.M{"$in": bson.A{orgIdentity(platform, org, ""), identity}},
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
		return false, err
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

func (this *client) ListIndividualSigning(opt dbmodels.IndividualSigningListOption) ([]dbmodels.IndividualSigningBasicInfo, error) {
	docFilter := docFilterOfIndividualSigning(opt.Platform, opt.OrgID, opt.RepoID)

	arrayFilter := bson.M{}
	if opt.CorporationEmail != "" {
		arrayFilter[fieldCorpID] = genCorpID(opt.CorporationEmail)
	}
	if opt.CLALanguage != "" {
		arrayFilter[fieldCLALanguage] = opt.CLALanguage
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
		return nil, err
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
