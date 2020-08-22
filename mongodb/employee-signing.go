package mongodb

import (
	"fmt"
	"strings"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zengchen1024/cla-server/dbmodels"
)

const fieldEmployeeSigningsID = "employees"

func emailToKey(email string) string {
	return strings.ReplaceAll(email, ".", "_")
}

func emailSuffixToKey(email string) string {
	return emailToKey(strings.Split(email, "@")[1])
}

func employeeSigningField(email string) string {
	return fmt.Sprintf("%s.%s", fieldEmployeeSigningsID, emailSuffixToKey(email))
}

func employeeSigningElemField(email string) func(string) string {
	return func(field string) string {
		return fmt.Sprintf("%s.%s", employeeSigningField(email), field)
	}
}

type employeeSignings struct {
	// Employees is the cla signing information of employee
	// key is the email of employee
	Employees map[string]employeeSigning `bson:"employees,omitempty"`
}

type employeeSigning struct {
	Name        string      `bson:"name"`
	Email       string      `bson:"email"`
	Enabled     bool        `bson:"enabled"`
	SigningInfo signingInfo `bson:"signing_info"`
}

func (c *client) SignAsEmployee(claOrgID string, info dbmodels.EmployeeSigningInfo) error {
	claOrg, err := c.GetCLAOrg(claOrgID)
	if err != nil {
		return err
	}

	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return fmt.Errorf("Failed to build body for signing as corporation, err:%v", err)
	}

	field := employeeSigningField(info.Email)

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{
			bson.M{"$match": bson.M{
				"platform": claOrg.Platform,
				"org_id":   claOrg.OrgID,
				"repo_id":  claOrg.RepoID,
				"apply_to": claOrg.ApplyTo,
				"enabled":  true,
			}},
			bson.M{"$project": bson.M{
				"count": bson.M{"$cond": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$%s", field)},
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", field),
						"cond":  bson.M{"$eq": bson.A{"$$this.email", info.Email}},
					}}},
					0,
				}},
			}},
		}

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		var count []struct {
			Count int `bson:"count"`
		}
		err = cursor.All(ctx, &count)
		if err != nil {
			return err
		}

		for _, item := range count {
			if item.Count != 0 {
				return fmt.Errorf("Failed to sign as employee, it has signed")
			}
		}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{field: bson.M(body)}})
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to sign as employee, the cla bound to org is not exist")
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to sign as employee, impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}
