package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func docFilterOfCorpManager(linkID string) bson.M {
	return bson.M{
		fieldLinkID:     linkID,
		fieldLinkStatus: linkStatusReady,
	}
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

func (this *client) AddCorporationManager(linkID string, opt []dbmodels.CorporationManagerCreateOption, managerNumber int) *dbmodels.DBError {
	toAdd := make(bson.A, 0, len(opt))
	emails := make(bson.A, 0, len(opt))
	for _, item := range opt {
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

		emails = append(emails, item.Email)
	}

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(
		fieldCorpManagers, false,
		bson.M{
			fieldCorpID: genCorpID(opt[0].Email),
			"email":     bson.M{"$in": emails},
		},
		docFilter,
	)

	f := func(ctx context.Context) *dbmodels.DBError {
		return this.pushArrayElems(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, toAdd,
		)
	}

	return withContextOfDB(f)
}

func (this *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (map[string]dbmodels.CorporationManagerCheckResult, *dbmodels.DBError) {
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
		return nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	result := map[string]dbmodels.CorporationManagerCheckResult{}
	for _, doc := range v {
		cm := doc.Managers
		if len(cm) == 0 {
			continue
		}

		item := &cm[0]
		orgRepo := parseOrgIdentity(doc.OrgIdentity)
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

func (this *client) ResetCorporationManagerPassword(linkID, email string, opt dbmodels.CorporationManagerResetPassword) *dbmodels.DBError {
	updateCmd := bson.M{
		"password": opt.NewPassword,
		"changed":  true,
	}

	elemFilter := elemFilterOfCorpManager(email)
	elemFilter["password"] = opt.OldPassword

	docFilter := docFilterOfCorpManager(linkID)
	arrayFilterByElemMatch(fieldCorpManagers, true, elemFilter, docFilter)

	f := func(ctx context.Context) *dbmodels.DBError {
		return this.updateArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, updateCmd)
	}

	return withContextOfDB(f)
}

func (this *client) ListCorporationManager(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, *dbmodels.DBError) {
	filterOfArray := filterOfCorpID(email)
	if role != "" {
		filterOfArray["role"] = role
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
			docFilterOfCorpManager(linkID), filterOfArray, project, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, systemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	ms := v[0].Managers
	if ms == nil {
		return nil, errNoChildDoc
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

func (this *client) DeleteCorporationManager(linkID string, emails []string) ([]dbmodels.CorporationManagerCreateOption, *dbmodels.DBError) {
	toDeleted := make(bson.A, 0, len(emails))
	for _, item := range emails {
		toDeleted = append(toDeleted, item)
	}

	elemFilter := bson.M{
		fieldCorpID: genCorpID(emails[0]),
		"email":     bson.M{"$in": toDeleted},
	}

	var v cCorpSigning
	f := func(ctx context.Context) *dbmodels.DBError {
		return this.pullAndReturnArrayElem(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilterOfCorpManager(linkID), elemFilter,
			&v,
		)
	}

	if err := withContextOfDB(f); err != nil {
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
