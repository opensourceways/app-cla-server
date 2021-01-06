package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
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

type corporationManagerDoc struct {
	ID               string `bson:"id" json:"id"`
	Name             string `bson:"name" json:"name" required:"true"`
	Role             string `bson:"role" json:"role" required:"true"`
	Email            string `bson:"email"  json:"email" required:"true"`
	Password         string `bson:"password" json:"password" required:"true"`
	InitialPWChanged bool   `bson:"changed" json:"changed"`
}

func corpManagerField(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpManagers, field)
}

func filterForCorpManager(filter bson.M) {
	filter["apply_to"] = dbmodels.ApplyToCorporation
	filter["enabled"] = true
	filter[fieldCorpManagers] = bson.M{"$type": "array"}
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
	body, err := structToMap1(info)
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
		return this.pushArrayElem1(
			ctx, this.corpSigningCollection, fieldCorpManagers, docFilter, body,
		)
	}

	return withContext1(f)
}

func managersToAdd(
	ctx context.Context, c *client, oid primitive.ObjectID,
	opt []dbmodels.CorporationManagerCreateOption, managerNumber int,
) ([]dbmodels.CorporationManagerCreateOption, error) {

	ms, err := c.listCorporationManager(ctx, oid, opt[0].Email, opt[0].Role)
	if err != nil {
		return nil, err
	}

	currentEmails := map[string]bool{}
	currentIDs := map[string]bool{}
	for _, item := range ms {
		currentEmails[item.Email] = true
		if item.ID != "" {
			currentIDs[item.ID] = true
		}
	}

	toAdd := make([]dbmodels.CorporationManagerCreateOption, 0, len(opt))
	for _, item := range opt {
		if _, ok := currentEmails[item.Email]; !ok {
			if item.ID == "" {
				toAdd = append(toAdd, item)
			} else {
				if _, ok = currentIDs[item.ID]; !ok {
					toAdd = append(toAdd, item)
				}
			}
		}
	}

	if len(ms)+len(toAdd) > managerNumber {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNumOfCorpManagersExceeded,
			Err:     fmt.Errorf("exceed %d managers allowed", managerNumber),
		}
	}

	return toAdd, nil
}

func (this *client) AddCorporationManager(orgCLAID string, opt []dbmodels.CorporationManagerCreateOption, managerNumber int) ([]dbmodels.CorporationManagerCreateOption, error) {
	oid, err := toObjectID(orgCLAID)
	if err != nil {
		return nil, err
	}

	var toAdd []dbmodels.CorporationManagerCreateOption

	f := func(ctx mongo.SessionContext) error {
		toAdd, err = managersToAdd(ctx, this, oid, opt, managerNumber)
		if err != nil {
			return err
		}
		if len(toAdd) == 0 {
			return nil
		}

		items := make(bson.A, 0, len(toAdd))
		for _, item := range toAdd {
			info := corporationManagerDoc{
				ID:       item.ID,
				Name:     item.Name,
				Email:    item.Email,
				Role:     item.Role,
				Password: item.Password,
			}

			body, err := structToMap(info)
			if err != nil {
				return err
			}
			addCorporationID(item.Email, body)

			items = append(items, body)
		}

		return this.pushArrayElems(
			ctx, this.orgCLACollection, fieldCorpManagers,
			filterOfDocID(oid), items,
		)
	}

	err = this.doTransaction(f)
	return toAdd, err
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
		return this.updateArrayElem1(
			ctx, this.corpSigningCollection, fieldCorpManagers,
			docFilter, elemFilter, updateCmd)
	}

	return withContext1(f)
}

func (this *client) listCorporationManager(ctx context.Context, orgCLAID primitive.ObjectID, email, role string) ([]corporationManagerDoc, error) {
	filterOfArray := filterOfCorpID(email)
	if role != "" {
		filterOfArray["role"] = role
	}

	project := bson.M{
		corpManagerField("id"):    1,
		corpManagerField("name"):  1,
		corpManagerField("email"): 1,
		corpManagerField("role"):  1,
	}

	var v []OrgCLA
	err := this.getArrayElem(
		ctx, this.orgCLACollection, fieldCorpManagers,
		filterOfDocID(orgCLAID), filterOfArray, project, &v,
	)
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: util.ErrNoDBRecord,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}
	return v[0].CorporationManagers, nil
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

func (this *client) DeleteCorporationManager(linkID string, emails []string) ([]dbmodels.CorporationManagerCreateOption, dbmodels.IDBError) {
	toDeleted := make(bson.A, 0, len(emails))
	for _, item := range emails {
		toDeleted = append(toDeleted, item)
	}

	elemFilter := bson.M{
		fieldCorpID: genCorpID(emails[0]),
		"email":     bson.M{"$in": toDeleted},
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
