package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func fieldNameOfCLA(applyTo string) string {
	if applyTo == dbmodels.ApplyToCorporation {
		return fieldCorpCLAs
	}
	return fieldIndividualCLAs
}

func docFilterOfCLA(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func elemFilterOfCLA(language string) bson.M {
	return bson.M{fieldCLALang: language}
}

func (this *client) HasCLA(linkID, applyTo, language string) (bool, error) {
	claField := fieldNameOfCLA(applyTo)

	project := bson.M{fmt.Sprintf("%s.url", claField): 1}

	var v []cLink
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.linkCollection, claField, docFilterOfCLA(linkID),
			elemFilterOfCLA(language), project, &v)
	}

	if err := withContext(f); err != nil {
		return false, err
	}

	if len(v) == 0 {
		return false, nil
	}

	doc := &v[0]
	if applyTo == dbmodels.ApplyToIndividual {
		if len(doc.IndividualCLAs) > 0 {
			return true, nil
		}
	} else {
		if len(doc.CorpCLAs) > 0 {
			return true, nil
		}
	}

	return false, nil
}

func (this *client) AddCLA(linkID, applyTo string, cla *dbmodels.CLACreateOption) *dbmodels.DBError {
	body, err := toDocOfCLA(cla)
	if err != nil {
		return err
	}

	claField := fieldNameOfCLA(applyTo)

	docFilter := docFilterOfCLA(linkID)
	arrayFilterByElemMatch(
		claField, false, elemFilterOfCLA(cla.Language), docFilter,
	)

	f := func(ctx context.Context) *dbmodels.DBError {
		return this.pushArrayElem(
			ctx, this.linkCollection, claField, docFilter, body,
		)
	}

	return withContextOfDB(f)
}

func (this *client) DeleteCLA(linkID, applyTo, language string) error {
	f := func(ctx context.Context) error {
		return this.pullArrayElem(
			ctx, this.linkCollection, fieldNameOfCLA(applyTo),
			docFilterOfCLA(linkID), elemFilterOfCLA(language),
		)
	}

	return withContext(f)
}

func (this *client) GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) (string, []dbmodels.CLADetail, error) {
	var project bson.M
	if applyTo == dbmodels.ApplyToIndividual {
		project = bson.M{fieldIndividualCLAs: 1}
	} else {
		project = bson.M{
			fieldIndividualCLAs: 0,
			fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
		}
	}

	var v cLink
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection, docFilterOfLink(orgRepo), project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return "", nil, err
	}

	if applyTo == dbmodels.ApplyToIndividual {
		return v.LinkID, toModelOfCLAs(v.IndividualCLAs), nil
	}
	return v.LinkID, toModelOfCLAs(v.CorpCLAs), nil
}

func (this *client) GetAllCLA(linkID string) (*dbmodels.CLAOfLink, error) {
	project := bson.M{
		fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
	}

	var v cLink
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection, docFilterOfCLA(linkID), project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return &dbmodels.CLAOfLink{
		IndividualCLAs: toModelOfCLAs(v.IndividualCLAs),
		CorpCLAs:       toModelOfCLAs(v.CorpCLAs),
	}, nil
}

func (this *client) GetCLAInfoToSign(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, error) {
	claField := fieldNameOfCLA(applyTo)

	fn := func(s string) string {
		return fmt.Sprintf("%s.%s", claField, s)
	}

	var v []cLink
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.linkCollection, claField,
			docFilterOfCLA(linkID), elemFilterOfCLA(claLang),
			bson.M{
				fn("fields"):             1,
				fn("cla_hash"):           1,
				fn("org_signature_hash"): 1,
			}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	}

	var doc []dCLA
	if applyTo == dbmodels.ApplyToIndividual {
		doc = v[0].IndividualCLAs
	} else {
		doc = v[0].CorpCLAs
	}

	if len(doc) == 0 {
		return nil, nil
	}

	item := &(doc[0])
	return &dbmodels.CLAInfo{
		CLAHash:          item.CLAHash,
		OrgSignatureHash: item.OrgSignatureHash,
		Fields:           toModelOfCLAFields(item.Fields),
	}, nil
}

func toModelOfCLAs(data []dCLA) []dbmodels.CLADetail {
	if data == nil {
		return nil
	}

	f := func(item *dCLA) *dbmodels.CLADetail {
		cla := dbmodels.CLADetail{
			Text:    item.Text,
			CLAHash: item.CLAHash,
		}

		cla.URL = item.URL
		cla.Language = item.Language

		if len(item.Fields) > 0 {
			fs := make([]dbmodels.Field, 0, len(item.Fields))
			for _, v := range item.Fields {
				fs = append(fs, dbmodels.Field{
					ID:          v.ID,
					Title:       v.Title,
					Type:        v.Type,
					Description: v.Description,
					Required:    v.Required,
				})
			}
			cla.Fields = fs
		}

		return &cla
	}

	r := make([]dbmodels.CLADetail, 0, len(data))
	for i := range data {
		r = append(r, *f(&data[i]))
	}
	return r
}

func toDocOfCLA(cla *dbmodels.CLACreateOption) (bson.M, *dbmodels.DBError) {
	info := &dCLA{
		URL:  cla.URL,
		Text: cla.Text,
		DCLAInfo: DCLAInfo{
			Fields:           toDocOfCLAField(cla.Fields),
			Language:         cla.Language,
			CLAHash:          cla.CLAHash,
			OrgSignatureHash: cla.OrgSignatureHash,
		},
	}
	r, err := structToMap(info)
	if err != nil {
		return nil, err
	}

	if cla.OrgSignature != nil {
		r[fieldOrgSignature] = *cla.OrgSignature
	}

	return r, nil
}

func toDocOfCLAField(fs []dbmodels.Field) []dField {
	if len(fs) == 0 {
		return nil
	}

	fields := make([]dField, 0, len(fs))
	for _, item := range fs {
		fields = append(fields, dField{
			ID:          item.ID,
			Title:       item.Title,
			Type:        item.Type,
			Description: item.Description,
			Required:    item.Required,
		})
	}
	return fields
}
