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

func filterForIndividualSigning(filter bson.M, advanced bool) {
	filter["apply_to"] = dbmodels.ApplyToIndividual
	filter["enabled"] = true
	if advanced {
		filter[fieldIndividuals] = bson.M{"$type": "array"}
	}
}

func filterOfDocForIndividualSigning(platform, org, repo string, advanced bool) bson.M {
	m, _ := filterOfOrgRepo(platform, org, repo)
	filterForIndividualSigning(m, advanced)
	return m
}

func (c *client) SignAsIndividual(orgCLAID, platform, org, repo string, info dbmodels.IndividualSigningInfo) error {
	oid, err := toObjectID(orgCLAID)
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
		notExist, err := c.isArrayElemNotExists(
			ctx, orgCLACollection, fieldIndividuals,
			filterOfDocForIndividualSigning(platform, org, repo, false),
			indexOfCorpManagerAndIndividual(info.Email),
		)
		if err != nil {
			return err
		}
		if !notExist {
			return dbmodels.DBError{
				ErrCode: util.ErrHasSigned,
				Err:     fmt.Errorf("he/she has signed"),
			}
		}

		return c.pushArrayElem(ctx, orgCLACollection, fieldIndividuals, filterOfDocID(oid), body)
	}

	return c.doTransaction(f)
}

func (c *client) DeleteIndividualSigning(platform, org, repo, email string) error {
	f := func(ctx mongo.SessionContext) error {
		return c.pullArrayElem(
			ctx, orgCLACollection, fieldIndividuals,
			filterOfDocForIndividualSigning(platform, org, repo, false),
			indexOfCorpManagerAndIndividual(email),
		)
	}

	// TODO don't use transaction if there is only one doc of cla org for individual signing
	return c.doTransaction(f)
}

func (c *client) UpdateIndividualSigning(platform, org, repo, email string, enabled bool) error {
	f := func(ctx mongo.SessionContext) error {
		return c.updateArrayElem(
			ctx, orgCLACollection, fieldIndividuals,
			filterOfDocForIndividualSigning(platform, org, repo, true),
			indexOfCorpManagerAndIndividual(email),
			bson.M{"enabled": enabled},
			true,
		)
	}

	return c.doTransaction(f)
}

func (c *client) IsIndividualSigned(platform, orgID, repoID, email string) (bool, error) {
	// must specify repo="" for filterOfOrgRepo to add org filter
	filterOfDoc, err := filterOfOrgRepo(platform, orgID, "")
	if err != nil {
		return false, err
	}
	filterForIndividualSigning(filterOfDoc, false)

	repo := dbValueOfRepo(orgID, repoID)
	if repo != "" {
		filterOfDoc[fieldRepo] = bson.M{"$in": bson.A{"", repo}}
	}

	var v []OrgCLA

	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, orgCLACollection, fieldIndividuals, filterOfDoc,
			indexOfCorpManagerAndIndividual(email),
			bson.M{
				fieldRepo:                         1,
				individualSigningField("enabled"): 1,
			},
			&v,
		)
	}

	if err := withContext(f); err != nil {
		return false, err
	}

	if len(v) == 0 {
		return false, nil
	}

	if repo != "" {
		bingo := false

		for i := 0; i < len(v); i++ {
			doc := &v[i]
			if doc.RepoID == repo {
				if !bingo {
					bingo = true
				}
				if len(doc.Individuals) > 0 {
					return doc.Individuals[0].Enabled, nil
				}
			}
		}
		if bingo {
			return false, nil
		}
	}

	for i := 0; i < len(v); i++ {
		doc := &v[i]
		if len(doc.Individuals) > 0 {
			return doc.Individuals[0].Enabled, nil
		}
	}

	return false, nil
}

func (c *client) ListIndividualSigning(opt dbmodels.IndividualSigningListOption) (map[string][]dbmodels.IndividualSigningBasicInfo, error) {
	filterOfDoc, err := filterOfOrgRepo(opt.Platform, opt.OrgID, opt.RepoID)
	if err != nil {
		return nil, err
	}
	if opt.CLALanguage != "" {
		filterOfDoc["cla_language"] = opt.CLALanguage
	}
	filterForIndividualSigning(filterOfDoc, true)

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

	var v []OrgCLA
	f := func(ctx context.Context) error {
		return c.getArrayElem(
			ctx, orgCLACollection, fieldIndividuals,
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
