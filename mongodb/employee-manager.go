package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func (this *client) AddEmployeeManager(linkID string, opt []dbmodels.CorporationManagerCreateOption) dbmodels.IDBError {
	toAdd := make(bson.A, 0, len(opt))
	for i := range opt {
		item := &opt[i]
		info := dCorpManager{
			ID:       item.ID,
			Name:     item.Name,
			Email:    item.Email,
			Role:     item.Role,
			Password: item.Password,
			CorpID:   genCorpID(item.Email),
		}

		body, err := structToMap(info)
		if err != nil {
			return err
		}

		toAdd = append(toAdd, body)
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElems(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), toAdd,
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
