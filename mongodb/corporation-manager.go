package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zengchen1024/cla-server/dbmodels"
	"github.com/zengchen1024/cla-server/models"
)

const fieldManagersID = "corporation_managers"

type corporationManager struct {
	Name          string `bson:"name"`
	Role          string `bson:"role"`
	Email         string `bson:"email"`
	Password      string `bson:"password"`
	CorporationID string `bson:"corporation_id"`
}

func corpoManagerKey(field string) string {
	return fmt.Sprintf("%s.%s", fieldManagersID, field)
}

func checkBeforeAddingCorporationManager(c *client, ctx mongo.SessionContext, claOrg models.CLAOrg, opt []dbmodels.CorporationManagerCreateOption) (int, int, error) {
	emails := make(bson.A, 0, len(opt))
	for _, item := range opt {
		emails = append(emails, item.Email)
	}

	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"platform": claOrg.Platform,
			"org_id":   claOrg.OrgID,
			"repo_id":  claOrg.RepoID,
			"apply_to": claOrg.ApplyTo,
			"enabled":  true,
		}},
		bson.M{"$project": bson.M{
			"role_count": bson.M{"$cond": bson.A{
				bson.M{"$isArray": fmt.Sprintf("$%s", fieldManagersID)},
				bson.M{"$size": bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldManagersID),
					"cond": bson.M{"$and": bson.A{
						bson.M{"$eq": bson.A{"$$this.corporation_id", opt[0].CorporationID}},
						bson.M{"$eq": bson.A{"$$this.role", opt[0].Role}},
					}},
				}}},
				0,
			}},
			"email_count": bson.M{"$cond": bson.A{
				bson.M{"$isArray": fmt.Sprintf("$%s", fieldManagersID)},
				bson.M{"$size": bson.M{"$filter": bson.M{
					"input": fmt.Sprintf("$%s", fieldManagersID),
					"cond":  bson.M{"$in": bson.A{"$$this.email", emails}},
				}}},
				0,
			}},
		}},
	}

	col := c.collection(claOrgCollection)
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}

	var count []struct {
		RoleCount  int `bson:"role_count"`
		EmailCount int `bson:"email_count"`
	}
	err = cursor.All(ctx, &count)
	if err != nil {
		return 0, 0, err
	}

	roleCount := 0
	emailCount := 0
	for _, item := range count {
		roleCount += item.RoleCount
		emailCount += item.EmailCount
	}
	return roleCount, emailCount, nil
}

func (c *client) AddCorporationManager(claOrgID string, opt []dbmodels.CorporationManagerCreateOption, managerNumber int) error {
	claOrg, err := c.GetCLAOrg(claOrgID)
	if err != nil {
		return err
	}

	updates := make(bson.A, 0, len(opt))
	for _, item := range opt {
		body, err := golangsdk.BuildRequestBody(item, "")
		if err != nil {
			return fmt.Errorf("Failed to build body for adding corporation manager, err:%v", err)
		}
		updates = append(updates, bson.M(body))
	}

	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx mongo.SessionContext) error {
		roleCount, emailCount, err := checkBeforeAddingCorporationManager(c, ctx, claOrg, opt)
		if err != nil {
			return fmt.Errorf("Failed to add corporation manager: check failed: %s", err.Error())
		}

		if roleCount+len(opt) > managerNumber {
			return fmt.Errorf("Failed to add corporation manager: it will exceed %d managers allowed", managerNumber)
		}
		if emailCount != 0 {
			return fmt.Errorf("Failed to add corporation manager: there are already %d same emails", emailCount)
		}

		col := c.collection(claOrgCollection)

		v, err := col.UpdateOne(
			ctx, bson.M{"_id": oid},
			bson.M{"$push": bson.M{fieldManagersID: bson.M{"$each": updates}}},
		)
		if err != nil {
			return fmt.Errorf("Failed to add corporation manager: add record failed: %s", err.Error())
		}

		if v.ModifiedCount != 1 {
			return fmt.Errorf("Failed to add corporation manager: impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) CheckCorporationManagerExist(opt dbmodels.CorporationManagerCheckInfo) (dbmodels.CorporationManagerCheckResult, error) {
	result := dbmodels.CorporationManagerCheckResult{}

	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return result, fmt.Errorf("Failed to build options to list corporation manager, err:%v", err)
	}
	filter := bson.M{
		corpoManagerKey("password"): opt.Password,
		"$or": bson.A{
			bson.M{corpoManagerKey("email"): opt.User},
			bson.M{corpoManagerKey("name"): opt.User},
		},
		"enabled": true,
	}
	for k, v := range body {
		filter[k] = v
	}

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				corpoManagerKey("role"):           1,
				corpoManagerKey("corporation_id"): 1,
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		return cursor.All(ctx, &v)
	}

	err = withContext(f)
	if err != nil {
		return result, err
	}

	if len(v) != 1 {
		return result, fmt.Errorf(
			"Failed to check corporation manager: there isn't only one cla orgonization binding which was signed by this corporation")
	}

	ms := v[0].CorporationManagers
	if ms == nil || len(ms) != 1 {
		return result, fmt.Errorf(
			"Failed to check corporation manager: there isn't only one corporation manager")
	}

	result.CorporationID = ms[0].CorporationID
	result.Role = ms[0].Role
	result.CLAOrgID = objectIDToUID(v[0].ID)
	return result, nil
}

func (c *client) ResetCorporationManagerPassword(opt dbmodels.CorporationManagerResetPassword) error {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return fmt.Errorf("Failed to build options to list corporation manager, err:%v", err)
	}
	filter := bson.M{
		corpoManagerKey("password"): opt.Password,
		"$or": bson.A{
			bson.M{corpoManagerKey("email"): opt.User},
			bson.M{corpoManagerKey("name"): opt.User},
		},
		"enabled": true,
	}
	for k, v := range body {
		filter[k] = v
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		update := bson.M{"$set": bson.M{fmt.Sprintf("%s.$.password", fieldManagersID): opt.NewPassword}}
		v, err := col.UpdateOne(ctx, filter, update)

		if err != nil {
			return fmt.Errorf("Failed to reset password for corporation manager: %s", err.Error())
		}

		if v.MatchedCount == 0 {
			return fmt.Errorf("Failed to reset password for corporation manager: user name or old password is not correct.")
		}

		if v.ModifiedCount != 1 {
			return fmt.Errorf("Failed to reset password for corporation manager: impossible.")
		}

		return nil
	}

	return withContext(f)
}
