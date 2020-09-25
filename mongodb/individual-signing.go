package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func individualSigningKey(email string) string {
	return fmt.Sprintf("%s.%s", fieldIndividuals, strings.ReplaceAll(email, ".", "_"))
}

type individualSigning struct {
	Name        string                   `bson:"name" json:"name" required:"true"`
	Email       string                   `bson:"email" json:"email" required:"true"`
	Enabled     bool                     `bson:"enabled" json:"enabled"`
	Date        string                   `bson:"date" json:"date" required:"true"`
	SigningInfo dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`
}

func additionalConditionForIndividualSigningDoc1(filter bson.M) {
	filter["apply_to"] = dbmodels.ApplyToIndividual
	filter["enabled"] = true
	filter[fieldIndividuals] = bson.M{"$type": "array"}
}

func (c *client) SignAsIndividual(claOrgID string, info dbmodels.IndividualSigningInfo) error {
	claOrg, err := c.GetBindingBetweenCLAAndOrg(claOrgID)
	if err != nil {
		return err
	}

	oid, _ := toObjectID(claOrgID)

	signing := individualSigning{
		Email:       info.Email,
		Name:        info.Name,
		Enabled:     info.Enabled,
		Date:        info.Date,
		SigningInfo: info.Info,
	}
	body, err := structToMap(signing)
	if err != nil {
		return err
	}
	addCorporationID(info.Email, body)

	f := func(ctx mongo.SessionContext) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{
			"platform": claOrg.Platform,
			"org_id":   claOrg.OrgID,
			"repo_id":  claOrg.RepoID,
		}
		additionalConditionForIndividualSigningDoc1(filter)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				"count": bson.M{"$cond": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$%s", fieldIndividuals)},
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", fieldIndividuals),
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
				return dbmodels.DBError{
					ErrCode: dbmodels.ErrHasSigned,
					Err:     fmt.Errorf("he/she has signed"),
				}
			}
		}

		r, err := col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$push": bson.M{fieldIndividuals: bson.M(body)}})
		if err != nil {
			return err
		}

		if r.ModifiedCount == 0 {
			return fmt.Errorf("impossible")
		}
		return nil
	}

	return c.doTransaction(f)
}

func (c *client) IsIndividualSigned(platform, orgID, repoID, email string) (bool, error) {
	opt := struct {
		Platform string `json:"platform" required:"true"`
		OrgID    string `json:"org_id" required:"true"`
		RepoID   string `json:"-" required:"true"`
		Email    string `json:"-" required:"true"`
	}{
		Platform: platform,
		OrgID:    orgID,
		RepoID:   repoID,
		Email:    email,
	}

	body, err := structToMap(opt)
	if err != nil {
		return false, err
	}

	signed := false

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M(body)
		filter[fieldRepo] = bson.M{"$in": bson.A{"", repoID}}
		additionalConditionForIndividualSigningDoc1(filter)

		pipeline := bson.A{
			bson.M{"$match": filter},
			bson.M{"$project": bson.M{
				fieldRepo: 1,
				"count": bson.M{"$cond": bson.A{
					bson.M{"$isArray": fmt.Sprintf("$%s", fieldIndividuals)},
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", fieldIndividuals),
						"cond": bson.M{"$and": bson.A{
							bson.M{"$eq": bson.A{"$$this.email", email}},
							bson.M{"$eq": bson.A{"$$this.enabled", true}},
						}},
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
			RepoID string `bson:"repo_id"`
			Count  int    `bson:"count"`
		}
		err = cursor.All(ctx, &count)
		if err != nil {
			return err
		}

		if repoID != "" {
			bingo := false

			for _, item := range count {
				if item.RepoID == repoID {
					if !bingo {
						bingo = true
					}

					if item.Count != 0 {
						signed = true
						return nil
					}
				}
			}
			if bingo {
				return nil
			}
		}

		for _, item := range count {
			if item.Count != 0 {
				signed = true
				return nil
			}
		}
		return nil
	}

	return signed, withContext(f)
}
