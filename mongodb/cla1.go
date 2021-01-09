package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCLA(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
}

func (this *client) GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) (string, []dbmodels.CLADetail, dbmodels.IDBError) {
	var project bson.M
	if applyTo == dbmodels.ApplyToIndividual {
		project = bson.M{
			fieldLinkID:         1,
			fieldIndividualCLAs: 1,
		}
	} else {
		project = bson.M{
			fieldOrgEmail:       0,
			fieldIndividualCLAs: 0,
			fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
		}
	}

	var v cLink
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc1(
			ctx, this.linkCollection, docFilterOfLink(orgRepo), project, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return "", nil, err
	}

	if applyTo == dbmodels.ApplyToIndividual {
		return v.LinkID, toModelOfCLAs(v.IndividualCLAs), nil
	}
	return v.LinkID, toModelOfCLAs(v.CorpCLAs), nil
}

func (this *client) GetAllCLA(linkID string) (*dbmodels.CLAOfLink, dbmodels.IDBError) {
	project := bson.M{
		fieldOrgEmail: 0,
		fmt.Sprintf("%s.%s", fieldCorpCLAs, fieldOrgSignature): 0,
	}

	var v cLink
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.getDoc1(
			ctx, this.linkCollection, docFilterOfCLA(linkID), project, &v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	return &dbmodels.CLAOfLink{
		IndividualCLAs: toModelOfCLAs(v.IndividualCLAs),
		CorpCLAs:       toModelOfCLAs(v.CorpCLAs),
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
