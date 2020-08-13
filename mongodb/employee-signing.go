package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/zengchen1024/cla-server/models"
)

type employeeSignings struct {
	// Employees is the cla signing information of employee
	// key is the email of employee
	Employees map[string]employeeSigning `bson:"employees,omitempty"`
}

type employeeSigning struct {
	SigningInfo signingInfo `bson:"signing_info"`
	Enabled     bool        `bson:"enabled"`
}

func emailToKey(email string) string {
	return strings.ReplaceAll(email, ".", "_")
}

func emailSuffixToKey(email string) string {
	return emailToKey(strings.Split(email, "@")[1])
}

func employeeSigningKey(email string) string {
	return fmt.Sprintf("employees.%s.%s", emailSuffixToKey(email), emailToKey(email))
}

func (c *client) SignAsEmployee(info models.EmployeeSigning) error {
	oid, err := toObjectID(info.CLAOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		k := employeeSigningKey(info.Email)
		info.Info["email"] = info.Email
		v := bson.M{k: bson.M{"signing_info": info.Info, "enabled": false}}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid, k: bson.M{"$exists": false}}, bson.M{"$set": v})
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return fmt.Errorf("Failed to add info when signing as employee, maybe he/she has signed")
		}
		return nil
	}

	return withContext(f)
}
