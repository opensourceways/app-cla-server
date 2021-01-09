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
	doc, err := structToMap1(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc1(ctx, this.individualSigningCollection, docFilter, doc)
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
	doc, err := structToMap1(data)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc1(ctx, this.corpSigningCollection, docFilter, doc)
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
		return this.pullArrayElem1(
			ctx, this.collectionOfSigning(applyTo), fieldCLAInfos,
			docFilterOfSigning(linkID), elemFilterOfCLA(claLang),
		)
	}

	return withContext1(f)
}

func (this *client) AddCLAInfo(linkID, applyTo string, info *dbmodels.CLAInfo) dbmodels.IDBError {
	doc, err := structToMap1(toDocOfCLAInfo(info))
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	arrayFilterByElemMatch(fieldCLAInfos, false, elemFilterOfCLA(info.CLALang), docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem1(
			ctx, this.collectionOfSigning(applyTo), fieldCLAInfos, docFilter, doc,
		)
	}

	return withContext1(f)
}

func toDocOfCLAInfo(info *dbmodels.CLAInfo) *DCLAInfo {
	return &DCLAInfo{
		Language:         info.CLALang,
		CLAHash:          info.CLAHash,
		OrgSignatureHash: info.OrgSignatureHash,
		Fields:           toDocOfCLAField(info.Fields),
	}
}
