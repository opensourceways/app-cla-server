package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func toDocOfCLAInfo(info *dbmodels.CLAInfo) *DCLAInfo {
	return &DCLAInfo{
		Language:         info.CLALang,
		CLAHash:          info.CLAHash,
		OrgSignatureHash: info.OrgSignatureHash,
		Fields:           toDocOfCLAField(info.Fields),
	}
}

func (this *client) InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo, claInfo *dbmodels.CLAInfo) *dbmodels.DBError {
	docFilter := bson.M{
		fieldOrgIdentity: info.String(),
		fieldLinkStatus:  linkStatusReady,
	}

	data := cCorpSigning{
		LinkID:      linkID,
		LinkStatus:  linkStatusReady,
		OrgIdentity: info.String(),
		OrgEmail:    info.OrgEmail,
		OrgAlias:    info.OrgAlias,
		CLAInfos:    []DCLAInfo{*toDocOfCLAInfo(claInfo)},
	}
	doc, err := structToMap(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) *dbmodels.DBError {
		_, err := this.replaceDoc(ctx, this.corpSigningCollection, docFilter, doc)
		return err
	}

	return withContextOfDB(f)
}

func (this *client) InitializeIndividualSigning(linkID string, orgRepo *dbmodels.OrgRepo, claInfo *dbmodels.CLAInfo) *dbmodels.DBError {
	docFilter := bson.M{
		fieldOrgIdentity: orgRepo.String(),
		fieldLinkStatus:  linkStatusReady,
	}

	data := cIndividualSigning{
		LinkID:      linkID,
		LinkStatus:  linkStatusReady,
		OrgIdentity: orgRepo.String(),
		CLAInfos:    []DCLAInfo{*toDocOfCLAInfo(claInfo)},
	}
	doc, err := structToMap(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) *dbmodels.DBError {
		_, err := this.replaceDoc(ctx, this.individualSigningCollection, docFilter, doc)
		return err
	}

	return withContextOfDB(f)
}

func (this *client) AddCLAInfo(linkID, applyTo string, info *dbmodels.CLAInfo) error {
	// TODO maybe need pull and push
	return nil
}

func (this *client) GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, *dbmodels.DBError) {
	elemFilter := elemFilterOfCLA(claLang)

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	col := this.individualSigningCollection
	if applyTo == dbmodels.ApplyToCorporation {
		col = this.corpSigningCollection
	}

	var v []struct {
		CLAInfos []DCLAInfo `bson:"cla_infos" json:"cla_infos"`
	}

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, col, fieldSingingCLAInfo, docFilter, elemFilter, nil, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, systemError(err)
	}

	if len(v) == 0 || len(v[0].CLAInfos) == 0 {
		return nil, nil
	}

	doc := &(v[0].CLAInfos[0])
	return &dbmodels.CLAInfo{
		CLAHash:          doc.CLAHash,
		OrgSignatureHash: doc.OrgSignatureHash,
		Fields:           toModelOfCLAFields(doc.Fields),
	}, nil
}
