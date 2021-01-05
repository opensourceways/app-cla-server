package mongodb

import (
	"context"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) InitializeIndividualSigning(linkID string) dbmodels.IDBError {
	docFilter := docFilterOfSigning(linkID)

	f := func(ctx context.Context) dbmodels.IDBError {
		_, err := this.replaceDoc1(ctx, this.individualSigningCollection, docFilter, docFilter)
		return err
	}

	return withContext1(f)
}

func (this *client) InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo) dbmodels.IDBError {
	docFilter := docFilterOfSigning(linkID)

	data := cCorpSigning{
		LinkID:      linkID,
		LinkStatus:  linkStatusReady,
		OrgIdentity: info.OrgRepoID(),
		OrgEmail:    info.OrgEmail,
		OrgAlias:    info.OrgAlias,
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
