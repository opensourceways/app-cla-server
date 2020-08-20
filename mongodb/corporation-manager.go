package mongodb

import (
	"context"
	"fmt"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/zengchen1024/cla-server/dbmodels"
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
	return fmt.Sprintf("%s.%s", corporationsID, field)
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
