package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCorpManager(linkID string) bson.M {
	return docFilterOfSigning(linkID)
}

func elemFilterOfCorpManager(email string) bson.M {
	return bson.M{
		fieldCorpID: genCorpID(email),
		fieldEmail:  email,
	}
}

func memberNameOfCorpManager(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpManagers, field)
}

func (this *client) AddCorpAdministrator(linkID string, opt *dbmodels.CorporationManagerCreateOption) dbmodels.IDBError {
	info := dCorpManager{
		ID:       opt.ID,
		Name:     opt.Name,
		Email:    opt.Email,
		Role:     dbmodels.RoleAdmin,
		Password: opt.Password,
		CorpID:   genCorpID(opt.Email),
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(
		fieldCorpManagers, false,
		bson.M{
			fieldCorpID: genCorpID(opt.Email),
			fieldRole:   dbmodels.RoleAdmin,
		},
		docFilter,
	)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, body,
		)
	}

	return withContext1(f)
}

func (this *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (map[string]dbmodels.CorporationManagerCheckResult, dbmodels.IDBError) {
	docFilter := bson.M{
		fieldLinkStatus:   linkStatusReady,
		fieldCorpManagers: bson.M{"$type": "array"},
		fieldLinkID:       opt.LinkID,
	}

	var elemFilter bson.M
	if opt.Email != "" {
		elemFilter = elemFilterOfCorpManager(opt.Email)
	} else {
		elemFilter = bson.M{
			fieldCorpID: opt.EmailSuffix,
			fieldID:     opt.ID,
		}
	}
	elemFilter[fieldPassword] = opt.Password

	project := bson.M{
		fieldLinkID:                           1,
		fieldOrgIdentity:                      1,
		fieldOrgEmail:                         1,
		fieldOrgAlias:                         1,
		memberNameOfCorpManager(fieldRole):    1,
		memberNameOfCorpManager(fieldName):    1,
		memberNameOfCorpManager(fieldEmail):   1,
		memberNameOfCorpManager(fieldChanged): 1,
		memberNameOfSignings(fieldCorp):       1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.corpSigningCollection, docFilter, project,
			map[string]func() bson.M{
				fieldCorpManagers: func() bson.M {
					return conditionTofilterArray(elemFilter)
				},
				fieldSignings: func() bson.M {
					return bson.M{"$and": bson.A{
						bson.M{"$isArray": fmt.Sprintf("$$this.%s", fieldDomains)},
						bson.M{"$in": bson.A{elemFilter[fieldCorpID], fmt.Sprintf("$$this.%s", fieldDomains)}},
					}}
				},
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}
	if len(v) == 0 {
		return nil, nil
	}

	result := map[string]dbmodels.CorporationManagerCheckResult{}
	for i := range v {
		doc := &v[i]
		cm := doc.Managers
		if len(cm) == 0 {
			continue
		}

		ss := doc.Signings
		if len(ss) == 0 {
			continue
		}

		item := &cm[0]
		orgRepo := dbmodels.ParseToOrgRepo(doc.OrgIdentity)
		result[doc.LinkID] = dbmodels.CorporationManagerCheckResult{
			Corp:             ss[0].CorpName,
			Name:             item.Name,
			Email:            item.Email,
			Role:             item.Role,
			InitialPWChanged: item.InitialPWChanged,

			OrgInfo: dbmodels.OrgInfo{
				OrgRepo: dbmodels.OrgRepo{
					Platform: orgRepo.Platform,
					OrgID:    orgRepo.OrgID,
					RepoID:   orgRepo.RepoID,
				},
				OrgEmail: doc.OrgEmail,
				OrgAlias: doc.OrgAlias,
			},
		}

	}
	return result, nil
}

func (this *client) ResetCorporationManagerPassword(linkID, email string, opt dbmodels.CorporationManagerResetPassword) dbmodels.IDBError {
	updateCmd := bson.M{
		fieldPassword: opt.NewPassword,
		fieldChanged:  true,
	}

	elemFilter := elemFilterOfCorpManager(email)
	elemFilter[fieldPassword] = opt.OldPassword

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(fieldCorpManagers, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, updateCmd)
	}

	return withContext1(f)
}

func (this *client) ListCorporationManager(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, dbmodels.IDBError) {
	domains, err := this.GetCorpEmailDomains(linkID, email)
	if err != nil {
		return nil, err
	}
	if domains == nil {
		return nil, nil
	}

	project := bson.M{
		memberNameOfCorpManager(fieldID):    1,
		memberNameOfCorpManager(fieldName):  1,
		memberNameOfCorpManager(fieldEmail): 1,
		memberNameOfCorpManager(fieldRole):  1,
	}

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.corpSigningCollection, docFilterOfCorpManager(linkID), project,
			map[string]func() bson.M{
				fieldCorpManagers: func() bson.M {
					c := bson.M{"$in": bson.A{fmt.Sprintf("$$this.%s", fieldCorpID), domains}}
					if role == "" {
						return c
					}

					return bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$$this." + fieldRole, role}},
						c,
					}}
				},
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	ms := v[0].Managers
	if ms == nil {
		return nil, nil
	}

	r := make([]dbmodels.CorporationManagerListResult, 0, len(ms))
	for i := range ms {
		item := &ms[i]
		r = append(r, dbmodels.CorporationManagerListResult{
			ID:    item.ID,
			Name:  item.Name,
			Email: item.Email,
			Role:  item.Role,
		})
	}
	return r, nil
}
