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
		fieldEmail: email,
	}
}

func memberNameOfCorpManager(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpManagers, field)
}

func (this *client) AddCorpAdministrator(
	si *dbmodels.SigningIndex,
	opt *dbmodels.CorporationManagerCreateOption,
) dbmodels.IDBError {
	index := newSigningIndex(si)

	info := dCorpManager{
		ID:       opt.ID,
		Name:     opt.Name,
		Role:     dbmodels.RoleAdmin,
		Email:    opt.Email,
		CorpID:   genCorpID(opt.Email),
		Password: opt.Password,
		CorpSID:  si.SigningId,
	}
	body, err := structToMap(info)
	if err != nil {
		return err
	}

	docFilter := index.docFilterOfSigning()
	docFilter["$nor"] = bson.A{
		bson.M{fieldCorpManagers: bson.M{"$elemMatch": bson.M{
			fieldCorpSId: si.SigningId,
			fieldRole:    dbmodels.RoleAdmin,
		}}},
		bson.M{fieldCorpManagers + "." + fieldID: opt.ID},
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.pushArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, body,
		)
	}

	return withContext1(f)
}

func (this *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (
	map[string]dbmodels.CorporationManagerCheckResult, dbmodels.IDBError,
) {
	docFilter := bson.M{
		fieldLinkStatus:   linkStatusReady,
		fieldCorpManagers: bson.M{"$type": "array"},
		fieldLinkID:       opt.LinkID,
	}

	elemFilter := make(bson.M)
	if opt.Email != "" {
		elemFilter[fieldEmail] = opt.Email
	} else {
		elemFilter[fieldID] = opt.ID
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
						bson.M{"$in": bson.A{
							opt.EmailSuffix,
							fmt.Sprintf("$$this.%s", fieldDomains),
						}},
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
		item := &cm[0]

		corpName := ""
		for _, s := range doc.Signings {
			if s.ID == item.CorpSID {
				corpName = s.CorpName
			}
		}
		if corpName == "" {
			continue
		}

		orgRepo := dbmodels.ParseToOrgRepo(doc.OrgIdentity)
		result[doc.LinkID] = dbmodels.CorporationManagerCheckResult{
			Corp:             corpName,
			Name:             item.Name,
			Email:            item.Email,
			Role:             item.Role,
			SigningId:        item.CorpSID,
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
	if opt.OldPassword != "" {
		elemFilter[fieldPassword] = opt.OldPassword
	}

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(fieldCorpManagers, true, elemFilter, docFilter)

	f := func(ctx context.Context) dbmodels.IDBError {
		return this.updateArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, updateCmd)
	}

	return withContext1(f)
}

func (this *client) GetCorporationDetail(si *dbmodels.SigningIndex) (
	detail dbmodels.CorporationDetail, err dbmodels.IDBError,
) {
	project := bson.M{
		memberNameOfSignings(fieldDomains):  1,
		memberNameOfCorpManager(fieldID):    1,
		memberNameOfCorpManager(fieldName):  1,
		memberNameOfCorpManager(fieldEmail): 1,
		memberNameOfCorpManager(fieldRole):  1,
	}

	index := newSigningIndex(si)

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.corpSigningCollection, index.docFilterOfSigning(), project,
			map[string]func() bson.M{
				fieldSignings: func() bson.M {
					return conditionTofilterArray(index.idFilter())
				},

				fieldCorpManagers: func() bson.M {
					return conditionTofilterArray(index.signingIdFilter())
				},
			},
			&v,
		)
	}

	if err1 := withContext(f); err != nil {
		err = newSystemError(err1)

		return
	}

	if len(v) == 0 {
		err = errNoDBRecord

		return
	}

	if s := v[0].Signings; len(s) != 0 {
		detail.EmailDomains = s[0].Domains
	}

	ms := v[0].Managers
	if len(ms) == 0 {
		return
	}

	r := make([]dbmodels.CorporationManagerListResult, 0, len(ms))
	for i := range ms {
		item := &ms[i]
		m := dbmodels.CorporationManagerListResult{
			ID:    item.ID,
			Name:  item.Name,
			Email: item.Email,
			Role:  item.Role,
		}

		if item.Role == dbmodels.RoleManager {
			r = append(r, m)
		} else {
			detail.Admin = m
		}
	}

	detail.Managers = r

	return
}
