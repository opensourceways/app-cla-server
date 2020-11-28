package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfEnabledLink(orgRepo *dbmodels.OrgRepo) bson.M {
	return bson.M{
		fieldOrgIdentity: orgIdentity(orgRepo),
		fieldLinkStatus:  linkStatusEnabled,
	}
}

func (this *client) getOrgCLAOfCorpCLA(orgRepo *dbmodels.OrgRepo, language, signatureMd5 string, result interface{}) error {
	fieldOfOrgSignature := memberNameOfCorpCLA(fieldOrgSignature)
	pipeline := bson.A{
		bson.M{"$match": docFilterOfEnabledLink(orgRepo)},
		bson.M{"$project": bson.M{
			fieldOrgAlias: 1,
			fieldOrgEmail: 1,
			fieldCorpCLAs: arrayElemFilter(fieldCorpCLAs, elemFilterOfCLA(language)),
		}},
		bson.M{"$unwind": "$" + fieldCorpCLAs},
		bson.M{"$project": bson.M{
			fieldOrgAlias:                 1,
			fieldOrgEmail:                 1,
			memberNameOfCorpCLA("text"):   1,
			memberNameOfCorpCLA("fields"): 1,
			fieldOfOrgSignature: bson.M{"$cond": bson.M{
				"if":   bson.M{"$eq": bson.A{signatureMd5, "$" + memberNameOfCorpCLA(fieldOrgSignatureTag)}},
				"then": "$$REMOVE",
				"else": "$" + fieldOfOrgSignature,
			}},
		}},
	}

	f := func(ctx context.Context) error {
		col := this.collection(this.linkCollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, result)
	}

	return withContext(f)
}

func (this *client) GetOrgCLAWhenSigningAsCorp(orgRepo *dbmodels.OrgRepo, language, signatureMd5 string) (*dbmodels.OrgCLAForSigning, error) {
	var v []struct {
		OrgAlias string `bson:"org_alias" json:"org_alias"`
		OrgEmail string `bson:"org_email" json:"-"`
		CorpCLA  dCLA   `bson:"corp_clas"`
	}
	err := this.getOrgCLAOfCorpCLA(orgRepo, language, signatureMd5, &v)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrNoDBRecord,
			Err:     fmt.Errorf("no cla binding found"),
		}
	}

	doc := &v[0]
	r := &dbmodels.OrgCLAForSigning{
		OrgDetail: dbmodels.OrgDetail{
			OrgInfo: dbmodels.OrgInfo{
				OrgAlias: doc.OrgAlias,
			},
			OrgEmail: doc.OrgEmail,
		},
		CLAInfo: toModelOfCLA(&doc.CorpCLA),
	}
	return r, nil
}

func (this *client) GetOrgCLAWhenSigningAsIndividual(orgRepo *dbmodels.OrgRepo, language string) (*dbmodels.OrgCLAForSigning, error) {
	var v []cOrgCLA
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.linkCollection, fieldIndividualCLAs,
			docFilterOfEnabledLink(orgRepo),
			elemFilterOfCLA(language),
			bson.M{
				fieldOrgAlias: 1,
				fieldOrgEmail: 1,
				fmt.Sprintf("%s.%s", fieldIndividualCLAs, "fields"): 1,
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 || len(v[0].IndividualCLAs) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrNoDBRecord,
			Err:     fmt.Errorf("no cla binding found"),
		}
	}

	doc := &v[0]
	r := &dbmodels.OrgCLAForSigning{
		OrgDetail: dbmodels.OrgDetail{
			OrgInfo: dbmodels.OrgInfo{
				OrgAlias: doc.OrgAlias,
			},
			OrgEmail: doc.OrgEmail,
		},
		CLAInfo: toModelOfCLA(&doc.IndividualCLAs[0]),
	}
	return r, nil
}
