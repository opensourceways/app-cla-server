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

type individualSigningDoc struct {
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

func (c *client) SignAsIndividual(claOrgID, platform, org, repo string, info dbmodels.IndividualSigningInfo) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	signing := individualSigningDoc{
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
		_, err := c.isIndividualSigned(platform, org, repo, info.Email, ctx)
		if err != nil {
			if !isHasNotSigned(err) {
				return err
			}
		} else {
			return dbmodels.DBError{
				ErrCode: util.ErrHasSigned,
				Err:     fmt.Errorf("he/she has signed"),
			}
		}

		col := c.collection(claOrgCollection)
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

func (c *client) DeleteIndividualSigning(claOrgID, email string) error {
	oid, err := toObjectID(claOrgID)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		col := c.collection(claOrgCollection)

		filter := bson.M{"_id": oid}
		filterForIndividualSigning(filter)

		update := bson.M{"$pull": bson.M{fieldIndividuals: bson.M{
			fieldCorporationID: util.EmailSuffix(email),
			"email":            email,
		}}}

		r, err := col.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}

		if r.MatchedCount == 0 {
			return dbmodels.DBError{
				ErrCode: util.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		return nil
	}

	return withContext(f)
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
				ErrCode: util.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the cla"),
			}
		}

		if r.ModifiedCount == 0 {
			return dbmodels.DBError{
				ErrCode: util.ErrInvalidParameter,
				Err:     fmt.Errorf("can't find the corresponding signing info"),
			}
		}
		return nil
	}

	return withContext(f)
}

func (c *client) IsIndividualSigned(platform, orgID, repoID, email string) (bool, error) {
	r := false

	f := func(ctx context.Context) error {
		v, err := c.isIndividualSigned(platform, orgID, repoID, email, ctx)
		r = v
		return err

	}

	err := withContext(f)
	return r, err
}

func (c *client) isIndividualSigned(platform, orgID, repoID, email string, ctx context.Context) (bool, error) {
	filterOfSigning := bson.M{
		fieldIndividuals: bson.M{"$filter": bson.M{
			"input": fmt.Sprintf("$%s", fieldIndividuals),
			"cond": bson.M{"$and": bson.A{
				bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
				bson.M{"$eq": bson.A{"$$this.email", email}},
			}},
		}},
	}

	project := bson.M{
		individualSigningField("enabled"): 1,
	}

	claOrg, err := c.getSigningDetail(platform, orgID, repoID, dbmodels.ApplyToIndividual, filterOfSigning, project, ctx)
	if err != nil {
		return false, err
	}

	return claOrg.Individuals[0].Enabled, nil
}

func (c *client) isIndividualSigned1(platform, orgID, repoID, email string, ctx context.Context) (bool, error) {
	filter := bson.M{
		"platform": platform,
		"org_id":   orgID,
		"apply_to": dbmodels.ApplyToIndividual,
		"enabled":  true,
	}
	if repoID == "" {
		filter[fieldRepo] = ""
	} else {
		filter[fieldRepo] = bson.M{"$in": bson.A{"", repoID}}
	}

	pipeline := bson.A{
		bson.M{"$match": filter},
		bson.M{"$project": bson.M{
			fieldRepo: 1,
			fieldIndividuals: bson.M{"$filter": bson.M{
				"input": fmt.Sprintf("$%s", fieldIndividuals),
				"cond": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$$this.corp_id", util.EmailSuffix(email)}},
					bson.M{"$eq": bson.A{"$$this.email", email}},
				}},
			}},
		}},
		bson.M{"$project": bson.M{
			fieldRepo:                         1,
			individualSigningField("enabled"): 1,
		}},
	}

	var v []CLAOrg
	f := func() error {
		col := c.collection(claOrgCollection)

		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		err = cursor.All(ctx, &v)
		if err != nil {
			return err
		}

		return nil
	}
	if err := f(); err != nil {
		return false, err
	}

	if len(v) == 0 {
		return false, dbmodels.DBError{
			ErrCode: util.ErrNoCLABindingDoc,
			Err:     fmt.Errorf("no record for this org/repo: %s/%s/%s", platform, orgID, repoID),
		}
	}

	err := dbmodels.DBError{
		ErrCode: util.ErrHasNotSigned,
		Err:     fmt.Errorf("he/she has not signed"),
	}
	if repoID != "" {
		bingo := false

		for _, doc := range v {
			if doc.RepoID == repoID {
				if !bingo {
					bingo = true
				}
				if len(doc.Individuals) > 0 {
					return doc.Individuals[0].Enabled, nil
				}
			}
		}
		if bingo {
			return false, err
		}
	}

	for _, doc := range v {
		if len(doc.Individuals) > 0 {
			return doc.Individuals[0].Enabled, nil
		}
	}

	return false, err
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

func toDBModelIndividualSigningBasicInfo(item individualSigningDoc) dbmodels.IndividualSigningBasicInfo {
	return dbmodels.IndividualSigningBasicInfo{
		Email:   item.Email,
		Name:    item.Name,
		Enabled: item.Enabled,
		Date:    item.Date,
	}
}
