package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func docFilterOfCorpManager(orgRepo *dbmodels.OrgRepo) bson.M {
	return docFilterOfEnabledLink(orgRepo)
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

func (c *client) AddCorporationManager(orgRepo *dbmodels.OrgRepo, opt []dbmodels.CorporationManagerCreateOption, managerNumber int) error {
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

	docFilter := docFilterOfCorpManager(orgRepo)

	f := func(ctx mongo.SessionContext) error {
		num, err := c.countArray(
			ctx, c.corpManagerCollection, fieldCorpManagers, docFilter,
			bson.M{
				fieldCorpID: genCorpID(opt[0].Email),
				"role":      opt[0].Role,
			},
		)
		if err != nil {
			return err
		}
		if num+len(opt) > managerNumber {
			return dbmodels.DBError{
				ErrCode: util.ErrNumOfCorpManagersExceeded,
				Err:     fmt.Errorf("exceed %d managers allowed", managerNumber),
			}
		}

		arrayFilterByElemMatch(
			fieldCorpManagers, false,
			bson.M{
				fieldCorpID: genCorpID(opt[0].Email),
				"email":     bson.M{"$in": emails},
			},
			docFilter,
		)

		return c.pushArrayElems(
			ctx, c.corpManagerCollection, fieldCorpManagers, docFilter, toAdd,
		)
	}

	return c.doTransaction(f)
}

func (c *client) GetCorpManager(opt *dbmodels.CorporationManagerCheckInfo) ([]dbmodels.CorporationManagerCheckResult, error) {
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
		fieldOrgIdentity:                   1,
		fieldOrgEmail:                      1,
		fieldOrgAlias:                      1,
		memberNameOfCorpManager("role"):    1,
		memberNameOfCorpManager("name"):    1,
		memberNameOfCorpManager("email"):   1,
		memberNameOfCorpManager("changed"): 1,
	}

	var v []cCorpManager
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, c.corpManagerCollection, fieldCorpManagers,
			bson.M{fieldLinkStatus: linkStatusEnabled},
			elemFilter, project, &v,
		)
	}
	if err := withContext(f); err != nil {
		return nil, err
	}

	result := make([]dbmodels.CorporationManagerCheckResult, 0, len(v))
	for _, doc := range v {
		if len(doc.CorpManagers) == 0 {
			continue
		}

		item := &doc.CorpManagers[0]
		orgRepo := parseOrgIdentity(doc.OrgIdentity)
		r := dbmodels.CorporationManagerCheckResult{
			Name:             item.Name,
			Email:            item.Email,
			Role:             item.Role,
			InitialPWChanged: item.InitialPWChanged,
			Platform:         orgRepo.Platform,
			OrgID:            orgRepo.OrgID,
			RepoID:           orgRepo.RepoID,
			OrgEmail:         doc.OrgEmail,
			OrgAlias:         doc.OrgAlias,
		}
		result = append(result, r)
	}

	return result, nil
}

func (c *client) ResetCorporationManagerPassword(orgRepo *dbmodels.OrgRepo, email string, opt *dbmodels.CorporationManagerResetPassword) error {
	elemFilter := elemFilterOfCorpManager(email)
	elemFilter["password"] = opt.OldPassword

	docFilter := docFilterOfCorpManager(orgRepo)
	arrayFilterByElemMatch(fieldSignings, true, elemFilter, docFilter)

	f := func(ctx context.Context) error {
		return c.updateArrayElem(
			ctx, c.corpManagerCollection, fieldCorpManagers, docFilter, elemFilter,
			bson.M{
				"password": opt.NewPassword,
				"changed":  true,
			}, false,
		)
	}

	return withContext(f)
}

func (c *client) ListCorporationManager(orgRepo *dbmodels.OrgRepo, email, role string) ([]dbmodels.CorporationManagerListResult, error) {
	docFilter := docFilterOfCorpManager(orgRepo)

	elemFilter := bson.M{
		fieldCorpID: genCorpID(email),
		"role":      role,
	}

	project := bson.M{
		memberNameOfCorpManager("id"):    1,
		memberNameOfCorpManager("name"):  1,
		memberNameOfCorpManager("email"): 1,
		memberNameOfCorpManager("role"):  1,
	}

	var v []cCorpManager
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, c.corpManagerCollection, fieldCorpManagers,
			docFilter, elemFilter, project, &v,
		)
	}
	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNoDBRecord,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}

	ms := v[0].CorpManagers
	r := make([]dbmodels.CorporationManagerListResult, 0, len(ms))
	for _, item := range ms {
		r = append(r, dbmodels.CorporationManagerListResult{
			ID:    item.ID,
			Name:  item.Name,
			Email: item.Email,
			Role:  item.Role,
		})
	}
	return r, nil
}

func (c *client) DeleteCorporationManager(orgRepo *dbmodels.OrgRepo, emails []string) ([]dbmodels.CorporationManagerCreateOption, error) {
	toDeleted := make(bson.A, 0, len(emails))
	for _, item := range emails {
		toDeleted = append(toDeleted, item)
	}

	elemFilter := bson.M{
		fieldCorpID: genCorpID(emails[0]),
		"email":     bson.M{"$in": toDeleted},
	}

	var v cCorpManager
	f := func(ctx context.Context) error {
		return c.pullAndReturnArrayElem(
			ctx, c.corpManagerCollection, fieldCorpManagers,
			docFilterOfCorpManager(orgRepo), elemFilter,
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	ms := v.CorpManagers
	deleted := make([]dbmodels.CorporationManagerCreateOption, 0, len(ms))
	for _, item := range ms {
		deleted = append(deleted, dbmodels.CorporationManagerCreateOption{
			Email: item.Email,
			Name:  item.Name,
		})
	}

	return deleted, nil
}
