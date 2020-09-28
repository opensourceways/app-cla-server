package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type corporationManagerDoc struct {
	Role     string `bson:"role" json:"role" required:"true"`
	Email    string `bson:"email"  json:"email" required:"true"`
	Password string `bson:"password" json:"password" required:"true"`
}

func corpManagerField(field string) string {
	return fmt.Sprintf("%s.%s", fieldCorpoManagers, field)
}

func filterForCorpManager(filter bson.M) {
	filter["apply_to"] = dbmodels.ApplyToCorporation
	filter["enabled"] = true
	filter[fieldCorpoManagers] = bson.M{"$type": "array"}
}

func (c *client) AddCorporationManager(claOrgID string, opt []dbmodels.CorporationManagerCreateOption, managerNumber int) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		ms, err := c.listCorporationManager(oid, opt[0].Email, opt[0].Role, ctx)
		if err != nil {
			return err
		}

		toAdd := map[string]dbmodels.CorporationManagerCreateOption{}
		for _, item := range opt {
			toAdd[item.Email] = item
		}
		for _, item := range ms {
			if _, ok := toAdd[item.Email]; ok {
				delete(toAdd, item.Email)
			}
		}

		if len(toAdd) == 0 {
			return nil
		}

		if len(ms)+len(toAdd) > managerNumber {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrNumOfCorpManagersExceeded,
				Err:     fmt.Errorf("exceed %d managers allowed", managerNumber),
			}
		}

		col := c.collection(claOrgCollection)

		updates := make(bson.A, 0, len(opt))
		for _, item := range toAdd {
			info := corporationManagerDoc{
				Email:    item.Email,
				Role:     item.Role,
				Password: item.Password,
			}

			body, err := structToMap(info)
			if err != nil {
				return err
			}
			addCorporationID(item.Email, body)

			updates = append(updates, bson.M(body))
		}

		v, err := col.UpdateOne(
			ctx, bson.M{"_id": oid},
			bson.M{"$push": bson.M{fieldCorpoManagers: bson.M{"$each": updates}}},
		)
		if err != nil {
			return fmt.Errorf("write db failed: %s", err.Error())
		}

		if v.ModifiedCount != 1 {
			return fmt.Errorf("impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (map[string][]dbmodels.CorporationManagerCheckResult, error) {
	filter := bson.M{}
	filterForCorpManager(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				"platform": 1,
				"org_id":   1,
				"repo_id":  1,
				fieldCorpoManagers: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorpoManagers),
					"cond": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(opt.User)}},
						bson.M{"$eq": bson.A{"$$this.email", opt.User}},
						bson.M{"$eq": bson.A{"$$this.password", opt.Password}},
					}},
				}}},
			},
			bson.M{"$project": bson.M{
				"platform":                1,
				"org_id":                  1,
				"repo_id":                 1,
				corpManagerField("role"):  1,
				corpManagerField("email"): 1,
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		return cursor.All(ctx, &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrInvalidParameter,
			Err:     fmt.Errorf("no manager found"),
		}
	}

	result := map[string][]dbmodels.CorporationManagerCheckResult{}
	for _, doc := range v {
		cm := doc.CorporationManagers
		if len(cm) == 0 {
			continue
		}

		// If len(cm) > 1, it happened when administrator add himself as the manager
		ms := make([]dbmodels.CorporationManagerCheckResult, 0, len(cm))
		for _, item := range cm {
			ms = append(ms, dbmodels.CorporationManagerCheckResult{
				Email:    item.Email,
				Role:     item.Role,
				Platform: doc.Platform,
				OrgID:    doc.OrgID,
				RepoID:   doc.RepoID,
			})
		}
		result[objectIDToUID(doc.ID)] = ms
	}
	return result, nil
}

func (c *client) ResetCorporationManagerPassword(claOrgID, email string, opt dbmodels.CorporationManagerResetPassword) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid}
	filterForCorpManager(filter)

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		update := bson.M{"$set": bson.M{fmt.Sprintf("%s.$[ms].password", fieldCorpoManagers): opt.NewPassword}}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"ms.corp_id":  util.EmailSuffix(email),
						"ms.email":    email,
						"ms.password": opt.OldPassword,
					},
				},
			},
		}

		v, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return err
		}

		if v.MatchedCount == 0 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		if v.ModifiedCount != 1 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("invalid email or old password"),
			}
		}

		return nil
	}

	return withContext(f)
}

func (c *client) listCorporationManager(claOrgID primitive.ObjectID, email, role string, ctx context.Context) ([]dbmodels.CorporationManagerListResult, error) {
	filter := bson.M{"_id": claOrgID}
	filterForCorpManager(filter)

	var v []CLAOrg

	cond := bson.A{
		bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
	}
	if role != "" {
		cond = append(cond, bson.M{"$eq": bson.A{"$$this.role", role}})

	}

	f := func() error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldCorpoManagers: bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldCorpoManagers),
					"cond":  cond,
				}}},
			},
			bson.M{"$project": bson.M{
				corpManagerField("email"): 1,
				corpManagerField("role"):  1,
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		return cursor.All(ctx, &v)
	}

	if err := f(); err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrInvalidParameter,
			Err:     fmt.Errorf("can't find the cla"),
		}
	}

	ms := v[0].CorporationManagers
	r := make([]dbmodels.CorporationManagerListResult, 0, len(ms))
	for _, item := range ms {
		r = append(r, dbmodels.CorporationManagerListResult{
			Email: item.Email,
			Role:  item.Role,
		})
	}
	return r, nil
}

func (c *client) ListCorporationManager(claOrgID, email, role string) ([]dbmodels.CorporationManagerListResult, error) {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return nil, err
	}

	var r []dbmodels.CorporationManagerListResult

	f := func(ctx context.Context) error {
		v, err := c.listCorporationManager(oid, email, role, ctx)
		r = v
		return err
	}

	err = withContext(f)

	return r, err
}

func (c *client) DeleteCorporationManager(claOrgID string, opt []dbmodels.CorporationManagerCreateOption) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		ms, err := c.listCorporationManager(oid, opt[0].Email, opt[0].Role, ctx)
		if err != nil {
			return err
		}

		all := map[string]bool{}
		for _, item := range ms {
			all[item.Email] = true
		}

		toDelete := make(bson.A, 0, len(opt))
		for _, item := range opt {
			if _, ok := all[item.Email]; ok {
				toDelete = append(toDelete, item.Email)
			}
		}
		if len(toDelete) == 0 {
			return nil
		}

		col := c.collection(claOrgCollection)

		update := bson.M{"$pull": bson.M{
			fieldCorpoManagers: bson.M{
				"email": bson.M{"$in": toDelete},
			},
		}}

		v, err := col.UpdateOne(ctx, bson.M{"_id": oid}, update)
		if err != nil {
			return fmt.Errorf("failed to write db: %s", err.Error())
		}

		if v.ModifiedCount != 1 {
			return fmt.Errorf("impossible.")
		}

		return nil
	}

	return c.doTransaction(f)
}
