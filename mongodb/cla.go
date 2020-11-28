package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func (this *client) GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) ([]dbmodels.CLA, error) {
	var project bson.M

	if applyTo == dbmodels.ApplyToIndividual {
		project = bson.M{fieldIndividualCLAs: 1}
	} else {
		project = bson.M{
			fieldIndividualCLAs: 0,
			fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
		}
	}

	var v cOrgCLA
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection,
			docFilterOfLink(orgRepo),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	r := toModelOfLinkCLA(&v)

	if applyTo == dbmodels.ApplyToIndividual {
		return r.IndividualCLAs, nil
	}
	return r.CorpCLAs, nil
}

func (this *client) GetAllCLA(orgRepo *dbmodels.OrgRepo) (*dbmodels.CLAOfLink, error) {
	var v cOrgCLA

	project := bson.M{
		fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
	}

	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.linkCollection,
			docFilterOfLink(orgRepo),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	return toModelOfLinkCLA(&v), nil
}

func (this *client) AddCLA(orgRepo *dbmodels.OrgRepo, applyTo string, cla *dbmodels.CLA) error {
	body, err := toDocOfCLA(cla)
	if err != nil {
		return err
	}

	claField := fieldNameOfCLA(applyTo)

	docFilter := docFilterOfLink(orgRepo)
	arrayFilterByElemMatch(
		claField, false, elemFilterOfCLA(cla.Language), docFilter,
	)

	f := func(ctx context.Context) error {
		return this.pushArrayElem(
			ctx, this.linkCollection, claField, docFilter, body,
		)
	}

	return withContext(f)
}

func (this *client) DeleteCLA(orgRepo *dbmodels.OrgRepo, applyTo, language string) error {
	claField := fieldNameOfCLA(applyTo)

	collection := this.individualSigningCollection
	if applyTo == dbmodels.ApplyToCorporation {
		collection = this.corpSigningCollection
	}

	f := func(ctx mongo.SessionContext) error {
		exist, err := this.isArrayElemNotExists(
			ctx, collection, fieldSignings,
			docFilterOfEnabledLink(orgRepo),
			bson.M{memberNameOfSignings(fieldCLALanguage): language},
		)
		if err != nil {
			return err
		}
		if exist {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrCLAHasBeenSigned,
				Err:     fmt.Errorf("cla has been signed"),
			}
		}
		return this.pullArrayElem(
			ctx, this.linkCollection, claField,
			docFilterOfLink(orgRepo), elemFilterOfCLA(language),
		)
	}

	return this.doTransaction(f)
}

func (this *client) DownloadOrgSignature(orgRepo *dbmodels.OrgRepo, language string) ([]byte, error) {
	elemFilter := elemFilterOfCLA(language)
	docFilter := docFilterOfLink(orgRepo)
	arrayFilterByElemMatch(fieldCorpCLAs, true, elemFilter, docFilter)

	var v []cOrgCLA
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.linkCollection, fieldCorpCLAs, docFilter, elemFilter,
			bson.M{memberNameOfCorpCLA(fieldOrgSignature): 1}, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, nil
	}
	return v[0].CorpCLAs[0].Text, nil
}

func toDocOfCLA(cla *dbmodels.CLA) (bson.M, error) {
	info := &dCLA{
		URL:      cla.URL,
		Text:     cla.Text,
		Language: cla.Language,
	}

	if len(cla.Fields) > 0 {
		fields := make([]dField, 0, len(cla.Fields))
		for _, item := range cla.Fields {
			fields = append(fields, dField{
				ID:          item.ID,
				Title:       item.Title,
				Type:        item.Type,
				Description: item.Description,
				Required:    item.Required,
			})
		}
		info.Fields = fields
	}

	if cla.OrgSignature != nil {
		info.Md5sumOfOrgSignature = util.Md5sumOfBytes(cla.OrgSignature)
	}

	r, err := structToMap(info)
	if err != nil {
		return nil, err
	}

	if cla.OrgSignature != nil {
		r[fieldOrgSignature] = cla.OrgSignature
	}
	return r, nil
}

func toModelOfCLA(cla *dCLA) *dbmodels.CLA {
	opt := &dbmodels.CLA{
		Text:         cla.Text,
		OrgSignature: cla.OrgSignature,
		CLAInfo: dbmodels.CLAInfo{
			URL:      cla.URL,
			Language: cla.Language,
		},
	}

	if len(cla.Fields) > 0 {
		fs := make([]dbmodels.Field, 0, len(cla.Fields))
		for _, v := range cla.Fields {
			fs = append(fs, dbmodels.Field{
				ID:          v.ID,
				Title:       v.Title,
				Type:        v.Type,
				Description: v.Description,
				Required:    v.Required,
			})
		}
		opt.Fields = fs
	}

	return opt
}

func fieldNameOfCLA(applyTo string) string {
	if applyTo == dbmodels.ApplyToCorporation {
		return fieldCorpCLAs
	}
	return fieldIndividualCLAs
}

func elemFilterOfCLA(language string) bson.M {
	return bson.M{fieldCLALanguage: language}
}

func toModelOfLinkCLA(doc *cOrgCLA) *dbmodels.CLAOfLink {
	convertCLAs := func(v []dCLA) []dbmodels.CLA {
		clas := make([]dbmodels.CLA, 0, len(v))
		for _, item := range v {
			clas = append(clas, *toModelOfCLA(&item))
		}

		return clas
	}

	r := dbmodels.CLAOfLink{}
	if len(doc.IndividualCLAs) > 0 {
		r.IndividualCLAs = convertCLAs(doc.IndividualCLAs)
	}

	if len(doc.CorpCLAs) > 0 {
		r.CorpCLAs = convertCLAs(doc.CorpCLAs)
	}
	return &r
}

func memberNameOfCorpCLA(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpCLAs, field)
}
