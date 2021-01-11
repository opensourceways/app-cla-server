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
		"email":     email,
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
			"role":      dbmodels.RoleAdmin,
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
	}

	var elemFilter bson.M
	if opt.Email != "" {
		elemFilter = elemFilterOfCorpManager(opt.Email)
	} else {
		elemFilter = bson.M{
			fieldCorpID: opt.EmailSuffix,
			"id":        opt.ID,
		}
	}
	elemFilter["password"] = opt.Password

	project := bson.M{
		fieldLinkID:                        1,
		fieldOrgIdentity:                   1,
		fieldOrgEmail:                      1,
		fieldOrgAlias:                      1,
		memberNameOfCorpManager("role"):    1,
		memberNameOfCorpManager("name"):    1,
		memberNameOfCorpManager("email"):   1,
		memberNameOfCorpManager("changed"): 1,
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}
	if len(v) == 0 {
		return nil, nil
	}

	result := map[string]dbmodels.CorporationManagerCheckResult{}
	for _, doc := range v {
		cm := doc.Managers
		if len(cm) == 0 {
			continue
		}

		item := &cm[0]
		orgRepo := dbmodels.ParseToOrgRepo(doc.OrgIdentity)
		result[doc.LinkID] = dbmodels.CorporationManagerCheckResult{
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
		"password": opt.NewPassword,
		"changed":  true,
	}

	elemFilter := elemFilterOfCorpManager(email)
	elemFilter["password"] = opt.OldPassword

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
	elemFilter := filterOfCorpID(email)
	if role != "" {
		elemFilter["role"] = role
	}

	project := bson.M{
		memberNameOfCorpManager("id"):    1,
		memberNameOfCorpManager("name"):  1,
		memberNameOfCorpManager("email"): 1,
		memberNameOfCorpManager("role"):  1,
	}

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), elemFilter, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord1
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
