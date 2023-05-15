package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) AddEmployeeManager(
	si *dbmodels.SigningIndex,
	opt *dbmodels.CorporationManagerCreateOption,
) dbmodels.IDBError {
	index := newSigningIndex(si)

	doc, err := structToMap(dCorpManager{
		ID:        opt.ID,
		Name:      opt.Name,
		Role:      opt.Role,
		Email:     opt.Email,
		CorpID:    genCorpID(opt.Email),
		Password:  opt.Password,
		SigningID: si.SigningId,
	})
	if err != nil {
		return err
	}

	docFilter := index.docFilterOfSigning()
	docFilter["$nor"] = bson.A{
		bson.M{fieldCorpManagers + "." + fieldEmail: opt.Email},
		bson.M{fieldCorpManagers + "." + fieldID: opt.ID},
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, doc,
		)
	}

	return withContext1(f)
}

func (this *client) DeleteEmployeeManager(linkID string, emails []string) ([]dbmodels.CorporationManagerCreateOption, dbmodels.IDBError) {
	elemFilter := bson.M{
		fieldEmail: bson.M{"$in": emails},
	}

	var v cCorpSigning
	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pullAndReturnArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), elemFilter,
			&v,
		)
	}

	if err := withContext1(f); err != nil {
		return nil, err
	}

	ms := v.Managers
	deleted := make([]dbmodels.CorporationManagerCreateOption, 0, len(ms))
	for _, item := range ms {
		deleted = append(deleted, dbmodels.CorporationManagerCreateOption{
			Email: item.Email,
			Name:  item.Name,
		})
	}

	return deleted, nil
}
