package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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
		_, err := c.isIndividualSigned(ctx, platform, org, repo, info.Email, false)
		if err == nil {
			return dbmodels.DBError{
				ErrCode: util.ErrHasSigned,
				Err:     fmt.Errorf("he/she has signed"),
			}
		}
		if !isErrorOfNotSigned(err) {
			return err
		}

		return c.pushArryItem(ctx, claOrgCollection, fieldIndividuals, filterOfDocID(oid), body)
	}

	return c.doTransaction(f)
}

func (c *client) DeleteIndividualSigning(platform, org, repo, email string) error {
	f := func(ctx mongo.SessionContext) error {
		claOrg, err := c.isIndividualSigned(ctx, platform, org, repo, email, false)
		if err != nil {
			if isErrorOfNotSigned(err) {
				return nil
			}
			return err
		}

		return c.pullArryItem(
			ctx, claOrgCollection, fieldIndividuals,
			filterOfDocID(claOrg.ID),
			bson.M{
				fieldCorporationID: genCorpID(email),
				"email":            email,
			},
		)
	}

	return c.doTransaction(f)
}

func (c *client) UpdateIndividualSigning(platform, org, repo, email string, enabled bool) error {
	f := func(ctx mongo.SessionContext) error {
		claOrg, err := c.isIndividualSigned(ctx, platform, org, repo, email, false)
		if err != nil {
			return err
		}

		return c.updateArryItem(
			ctx, claOrgCollection, fieldIndividuals,
			filterOfDocID(claOrg.ID),
			bson.M{
				fieldCorporationID: genCorpID(email),
				"email":            email,
				"enabled":          !enabled,
			},
			bson.M{"enabled": enabled},
			false,
		)
	}

	return c.doTransaction(f)
}

func (c *client) IsIndividualSigned(platform, orgID, repoID, email string) (bool, error) {
	r := false

	f := func(ctx context.Context) error {
		v, err := c.isIndividualSigned(ctx, platform, orgID, repoID, email, true)
		if err == nil {
			r = v.Individuals[0].Enabled
		}

		return err
	}

	err := withContext(f)
	return r, err
}

func (c *client) isIndividualSigned(ctx context.Context, platform, orgID, repoID, email string, orgCared bool) (*CLAOrg, error) {
	filterOfDoc := bson.M{
		"platform": platform,
		"org_id":   orgID,
		"apply_to": dbmodels.ApplyToIndividual,
		"enabled":  true,
	}
	if repoID != "" && orgCared {
		filterOfDoc[fieldRepo] = bson.M{"$in": bson.A{"", repoID}}
	} else {
		filterOfDoc[fieldRepo] = repoID
	}

	var v []CLAOrg

	err := c.getArrayElem(
		ctx, claOrgCollection, fieldIndividuals, filterOfDoc,
		bson.M{
			fieldCorporationID: genCorpID(email),
			"email":            email,
		},
		bson.M{
			fieldRepo:                         1,
			individualSigningField("enabled"): 1,
		},
		&v,
	)
	if err != nil {
		return nil, err
	}

	return c.getSigningDetail(
		platform, orgID, repoID, orgCared, v,
		func(doc *CLAOrg) bool {
			return len(doc.Individuals) > 0
		},
	)
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

	filterOfDoc, err := structToMap(info)
	if err != nil {
		return nil, err
	}
	filterForIndividualSigning(filterOfDoc)

	filterOfArray := bson.M{}
	if opt.CorporationEmail != "" {
		filterOfArray = filterOfCorpID(opt.CorporationEmail)
	}

	project := bson.M{
		individualSigningField("email"):   1,
		individualSigningField("name"):    1,
		individualSigningField("enabled"): 1,
		individualSigningField("date"):    1,
	}

	var v []CLAOrg
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, claOrgCollection, fieldIndividuals,
			filterOfDoc, filterOfArray, project, &v)
	}

	if err = withContext(f); err != nil {
		return nil, err
	}

	r := map[string][]dbmodels.IndividualSigningBasicInfo{}

	for i := 0; i < len(v); i++ {
		rs := v[i].Individuals
		if len(rs) == 0 {
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
