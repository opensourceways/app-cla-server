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

		email, err := this.encrypt.encryptStr(item.Email)
		if err != nil {
			return err
		}

		info := dCorpManager{
			ID:       item.ID,
			Name:     item.Name,
			Email:    email,
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
	encryptedEmails := make([]string, 0, len(emails))
	m := map[string]string{}
	for _, item := range emails {
		email, err := this.encrypt.encryptStr(item)
		if err != nil {
			return nil, err
		}
		encryptedEmails = append(encryptedEmails, email)
		m[email] = item
	}

	elemFilter := bson.M{
		fieldCorpID: genCorpID(emails[0]),
		fieldEmail:  bson.M{"$in": encryptedEmails},
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
			Email: m[item.Email],
			Name:  item.Name,
		})
	}

	return deleted, nil
}
