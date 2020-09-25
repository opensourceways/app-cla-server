package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type individualSigning struct {
	Name        string                   `bson:"name" json:"name" required:"true"`
	Email       string                   `bson:"email" json:"email" required:"true"`
	Enabled     bool                     `bson:"enabled" json:"enabled"`
	Date        string                   `bson:"date" json:"date" required:"true"`
	SigningInfo dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`
}

func individualSigningField(key string) string {
	return fmt.Sprintf("%s.%s", fieldIndividuals, key)
}

func filterForIndividualSigning(filter bson.M) {
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
		filterForIndividualSigning(filter)

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

func (c *client) UpdateIndividualSigning(claOrgID, email string, enabled bool) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForIndividualSigning(filter)

		update := bson.M{"$set": bson.M{fmt.Sprintf("%s.$[ms].enabled", fieldIndividuals): enabled}}

		updateOpt := options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: bson.A{
					bson.M{
						"ms.corp_id": util.EmailSuffix(email),
						"ms.email":   email,
						"ms.enabled": !enabled,
					},
				},
			},
		}

		r, err := col.UpdateOne(ctx, filter, update, &updateOpt)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		if r.ModifiedCount == 0 {
			return dbmodels.DBError{
				ErrCode: dbmodels.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the corresponding signing info"),
			}
		}
		return nil
	}

	return withContext(f)
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
		filterForIndividualSigning(filter)

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

func (c *client) ListIndividualSigning(opt dbmodels.IndividualSigningListOption) (map[string][]dbmodels.IndividualSigningBasicInfo, error) {
	info := struct {
		Platform    string `json:"platform" required:"true"`
		OrgID       string `json:"org_id" required:"true"`
		RepoID      string `json:"repo_id,omitempty"`
		CLALanguage string `json:"cla_language,omitempty"`
	}{
		Platform:    opt.Platform,
		OrgID:       opt.OrgID,
		RepoID:      opt.RepoID,
		CLALanguage: opt.CLALanguage,
	}

	body, err := structToMap(info)
	if err != nil {
		return nil, err
	}
	filter := bson.M(body)
	filterForIndividualSigning(filter)

	var v []CLAOrg

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		pipeline := bson.A{bson.M{"$match": filter}}

		if opt.CorporationEmail != "" {
			pipeline = append(
				pipeline,
				bson.M{"$project": bson.M{
					fieldIndividuals: bson.M{"$filter": bson.M{
						"input": fmt.Sprintf("$%s", fieldIndividuals),
						"cond":  bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(opt.CorporationEmail)}},
					}}},
				},
			)
		}

		pipeline = append(
			pipeline,
			bson.M{"$project": bson.M{
				individualSigningField("email"):   1,
				individualSigningField("name"):    1,
				individualSigningField("enabled"): 1,
				individualSigningField("date"):    1,
			}},
		)

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

	r := map[string][]dbmodels.IndividualSigningBasicInfo{}

	for i := 0; i < len(v); i++ {
		rs := v[i].Individuals
		if rs == nil || len(rs) == 0 {
			continue
		}

		es := make([]dbmodels.IndividualSigningBasicInfo, 0, len(rs))
		for _, item := range rs {
			es = append(es, toDBModelIndividualSigningBasicInfo(item))
		}
		r[objectIDToUID(v[i].ID)] = es
	}

	return r, nil
}

func toDBModelIndividualSigningBasicInfo(item individualSigning) dbmodels.IndividualSigningBasicInfo {
	return dbmodels.IndividualSigningBasicInfo{
		Email:   item.Email,
		Name:    item.Name,
		Enabled: item.Enabled,
		Date:    item.Date,
	}
}
