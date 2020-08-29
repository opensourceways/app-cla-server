package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/huaweicloud/golangsdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/zengchen1024/cla-server/dbmodels"
	"github.com/zengchen1024/cla-server/models"
)

const fieldEmployeeSigningsID = "employees"

func additionalConditionForIndividualSigningDoc(filter bson.M, email string) {
	filter["apply_to"] = models.ApplyToIndividual
	filter["enabled"] = true

	filter[employeeSigningField(email)] = bson.M{"$exists": true}
}

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
	claOrg, err := c.GetBindingBetweenCLAAndOrg(claOrgID)
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

		filter := bson.M{
			"platform": claOrg.Platform,
			"org_id":   claOrg.OrgID,
			"repo_id":  claOrg.RepoID,
		}
		additionalConditionForIndividualSigningDoc(filter, info.Email)

		pipeline := bson.A{
			bson.M{"$match": filter},
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

func (c *client) ListEmployeeSigning(opt dbmodels.EmployeeSigningListOption) (map[string][]dbmodels.EmployeeSigningInfo, error) {
	body, err := golangsdk.BuildRequestBody(opt, "")
	if err != nil {
		return nil, fmt.Errorf("build options to list employee signing failed, err:%v", err)
	}
	filter := bson.M(body)
	additionalConditionForIndividualSigningDoc(filter, opt.CorporationEmail)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		fieldFunc := employeeSigningElemField(opt.CorporationEmail)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldFunc("email"):   1,
				fieldFunc("name"):    1,
				fieldFunc("enabled"): 1,
			}},
		}
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("error find bindings: %v", err)
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return fmt.Errorf("error decoding to bson struct of employee signing: %v", err)
		}
		return nil
	}

	err = withContext(f)
	if err != nil {
		return nil, err
	}

	r := map[string][]dbmodels.EmployeeSigningInfo{}

	suffix := emailSuffixToKey(opt.CorporationEmail)

	for i := 0; i < len(v); i++ {
		m := v[i].Employees
		if m == nil || len(m) == 0 {
			continue
		}

		es, ok := m[suffix]
		if !ok || len(es) == 0 {
			continue
		}

		es1 := make([]dbmodels.EmployeeSigningInfo, 0, len(es))
		for _, item := range es {
			es1 = append(es1, toDBModelEmployeeSigningInfo(item))
		}
		r[objectIDToUID(v[i].ID)] = es1
	}

	return r, nil
}
func (c *client) UpdateEmployeeSigning(claOrgID, email string, opt dbmodels.EmployeeSigningUpdateInfo) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		additionalConditionForIndividualSigningDoc(filter, email)

		field := employeeSigningField(email)

		update := bson.M{"$set": bson.M{fmt.Sprintf("%s.$[ms].enabled", field): opt.Enabled}}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"ms.enabled": !opt.Enabled,
						"ms.email":   email,
					},
				},
			},
		}

		r, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return fmt.Errorf("Failed to update employee signing: %s", err.Error())
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to update employee signing, the cla which employee had signed is not exist")
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("Failed to update employee signing, impossible")
		}
		return nil
	}

	return withContext(f)
}

func toDBModelEmployeeSigningInfo(item employeeSigning) dbmodels.EmployeeSigningInfo {
	return dbmodels.EmployeeSigningInfo{
		Email:   item.Email,
		Name:    item.Name,
		Enabled: item.Enabled,
	}
}
