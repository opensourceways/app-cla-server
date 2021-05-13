package mongodb

import (
	"context"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) InitializeIndividualSigning(linkID string, claInfo *dbmodels.CLAInfo) dbmodels.IDBError {
	docFilter := docFilterOfSigning(linkID)

	data := cIndividualSigning{
		LinkID:     linkID,
		LinkStatus: linkStatusReady,
	}
	if claInfo != nil {
		data.CLAInfos = []DCLAInfo{*toDocOfCLAInfo(claInfo)}
	}
	doc, err := structToMap(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc(ctx, this.individualSigningCollection, docFilter, doc)
		return err
	}

	return withContext1(f)
}

func (this *client) InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo, claInfo *dbmodels.CLAInfo) dbmodels.IDBError {
	docFilter := docFilterOfSigning(linkID)

	data := cCorpSigning{
		LinkID:      linkID,
		LinkStatus:  linkStatusReady,
		OrgIdentity: info.OrgRepoID(),
		OrgEmail:    info.OrgEmail,
		OrgAlias:    info.OrgAlias,
	}
	if claInfo != nil {
		data.CLAInfos = []DCLAInfo{*toDocOfCLAInfo(claInfo)}
	}
	doc, err := structToMap(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc(ctx, this.corpSigningCollection, docFilter, doc)
		return err
	}

	return withContext1(f)
}

func (this *client) collectionOfSigning(applyTo string) string {
	if applyTo == dbmodels.ApplyToCorporation {
		return this.corpSigningCollection
	}
	return this.individualSigningCollection
}

func (this *client) DeleteCLAInfo(linkID, applyTo, claLang string) dbmodels.IDBError {
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pullArrayElem(
			ctx, this.collectionOfSigning(applyTo), fieldCLAInfos,
			docFilterOfSigning(linkID), elemFilterOfCLA(claLang),
		)
	}

	return withContext1(f)
}

func (this *client) AddCLAInfo(linkID, applyTo string, info *dbmodels.CLAInfo) dbmodels.IDBError {
	doc, err := structToMap(toDocOfCLAInfo(info))
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldCLAInfos, false, elemFilterOfCLA(info.CLALang), docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(
			ctx, this.collectionOfSigning(applyTo), fieldCLAInfos, docFilter, doc,
		)
	}

	return withContext1(f)
}

func (this *client) GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, dbmodels.IDBError) {
	elemFilter := elemFilterOfCLA(claLang)

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	var v []struct {
		CLAInfos []DCLAInfo `bson:"cla_infos" json:"cla_infos"`
	}

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.collectionOfSigning(applyTo), fieldCLAInfos, docFilter, elemFilter, nil, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	if len(v[0].CLAInfos) == 0 {
		return nil, nil
	}

	doc := &(v[0].CLAInfos[0])
	return &dbmodels.CLAInfo{
		CLAHash: doc.CLAHash,
		Fields:  toModelOfCLAFields(doc.Fields),
	}, nil
}

func toDocOfCLAInfo(info *dbmodels.CLAInfo) *DCLAInfo {
	return &DCLAInfo{
		Language: info.CLALang,
		CLAHash:  info.CLAHash,
		Fields:   toDocOfCLAField(info.Fields),
	}
}
