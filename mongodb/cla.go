package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func (this *client) GetCLAForSigning(orgRepo *dbmodels.OrgRepo, applyTo string) ([]dbmodels.CLA, error) {
	var project bson.M

	if applyTo == dbmodels.ApplyToIndividual {
		project = bson.M{fieldIndividualCLAs: 1}
	} else {
		project = bson.M{
			fieldOrgEmail:       0,
			fieldIndividualCLAs: 0,
			fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
		}
	}

	var v cOrgCLA
	f := func(ctx context.Context) error {
		return this.getDoc(
			ctx, this.orgCLACollection,
			docFilterOfLink(orgRepo),
			project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	r := toModelOfOrgCLA(&v)

	if applyTo == dbmodels.ApplyToIndividual {
		return r.IndividualCLAs, nil
	}
	return r.CorpCLAs, nil
}

func (this *client) AddCLA(orgRepo *dbmodels.OrgRepo, applyTo string, cla *dbmodels.CLA) error {
	body, err := toDocOfCLA(cla)
	if err != nil {
		return err
	}

	claField := fieldNameOfCLA(applyTo)

	docFilter := docFilterOfLink(orgRepo)
	arrayFilterByElemMatch(
		claField, false, docFilterOfCLA(cla.Language), docFilter,
	)

	f := func(ctx context.Context) error {
		return this.pushArrayElem(
			ctx, this.orgCLACollection, claField, docFilter, body,
		)
	}

	return withContext(f)
}

func (this *client) DeleteCLA(orgRepo *dbmodels.OrgRepo, applyTo, language string) error {
	claField := fieldNameOfCLA(applyTo)
	claFilter := docFilterOfCLA(language)

	docFilter := docFilterOfLink(orgRepo)
	// arrayFilterByElemMatch(claField, true, claFilter, docFilter)

	f := func(ctx context.Context) error {
		return this.pullArrayElem(
			ctx, this.orgCLACollection, claField, docFilter, claFilter,
		)
	}

	return withContext(f)
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

func docFilterOfCLA(language string) bson.M {
	return bson.M{fieldCLALanguage: language}
}
