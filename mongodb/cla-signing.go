package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func toDocOfCLAInfo(info *dbmodels.CLAInfo) (bson.M, error) {
	v := DCLAInfo{
		Language:         info.CLALang,
		CLAHash:          info.CLAHash,
		OrgSignatureHash: info.OrgSignatureHash,
		Fields:           toDocOfCLAField(info.Fields),
	}
	return structToMap(&v)
}

func (this *client) InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo, claInfo *dbmodels.CLAInfo) error {
	docFilter := bson.M{
		fieldOrgIdentity: orgIdentity(&info.OrgRepo),
		fieldLinkStatus:  linkStatusReady,
	}

	docOfCLAInfo, err := toDocOfCLAInfo(claInfo)
	if err != nil {
		return err
	}

	doc := bson.M{
		fieldLinkID:         linkID,
		fieldOrgEmail:       info.OrgEmail,
		fieldOrgAlias:       info.OrgAlias,
		fieldSignings:       "",
		fieldSingingCLAInfo: bson.A{docOfCLAInfo},
		fieldCorpManagers:   "",
	}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) error {
		_, err := this.replaceDoc(ctx, this.corpSigningCollection, docFilter, doc)
		return err
	}

	return withContext(f)
}

func (this *client) InitializeIndividualSigning(linkID string, orgRepo *dbmodels.OrgRepo, claInfo *dbmodels.CLAInfo) error {
	docFilter := bson.M{
		fieldOrgIdentity: orgIdentity(orgRepo),
		fieldLinkStatus:  linkStatusReady,
	}

	docOfCLAInfo, err := toDocOfCLAInfo(claInfo)
	if err != nil {
		return err
	}

	doc := bson.M{
		fieldLinkID:         linkID,
		fieldSignings:       "",
		fieldSingingCLAInfo: bson.A{docOfCLAInfo},
	}
	for k, v := range docFilter {
		doc[k] = v
	}

	f := func(ctx context.Context) error {
		_, err := this.replaceDoc(ctx, this.individualSigningCollection, docFilter, doc)
		return err
	}

	return withContext(f)
}

func (this *client) AddCLAInfo(linkID, applyTo string, info *dbmodels.CLAInfo) error {
	// TODO maybe need pull and push
	return nil
}

func (this *client) GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, error) {
	docFilter := docFilterOfIndividualSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilterOfCLA(claLang), docFilter)

	col := this.individualSigningCollection
	return this.getCLAInfo(col, claLang, &docFilter)
}

func (this *client) getCLAInfo(col, claLang string, docFilter *bson.M) (*dbmodels.CLAInfo, error) {
	var v []struct {
		CLAInfo []DCLAInfo `bson:"cla_info" json:"cla_info"`
	}

	elemFilter := elemFilterOfCLA(claLang)

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, col, fieldSingingCLAInfo, *docFilter, elemFilter, nil, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 || len(v[0].CLAInfo) == 0 {
		return nil, nil
	}

	doc := &(v[0].CLAInfo[0])
	return &dbmodels.CLAInfo{
		CLAHash:          doc.CLAHash,
		OrgSignatureHash: doc.OrgSignatureHash,
		Fields:           toModelOfCLAFields(doc.Fields),
	}, nil
}
