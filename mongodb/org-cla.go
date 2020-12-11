package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldIndividuals     = "individuals"
	fieldEmployees       = "employees"
	fieldCorporations    = "corporations"
	fieldCorpManagers    = "corp_managers"
	fieldCorporationID   = "corp_id"
	fieldOrgSignature    = "org_signature"
	fieldOrgSignatureTag = "md5sum"
	fieldRepo            = "repo_id"
)

func filterForClaOrgDoc(filter bson.M) {
	filter["enabled"] = true
}

type OrgCLA struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	CreatedAt   time.Time `bson:"created_at" json:"-"`
	UpdatedAt   time.Time `bson:"updated_at" json:"-"`
	Platform    string    `bson:"platform" json:"platform" required:"true"`
	OrgID       string    `bson:"org_id" json:"org_id" required:"true"`
	RepoID      string    `bson:"repo_id" json:"repo_id"`
	OrgAlias    string    `bson:"org_alias" json:"org_alias"`
	CLAID       string    `bson:"cla_id" json:"cla_id" required:"true"`
	CLALanguage string    `bson:"cla_language" json:"cla_language" required:"true"`
	ApplyTo     string    `bson:"apply_to" json:"apply_to" required:"true"`
	OrgEmail    string    `bson:"org_email" json:"org_email" required:"true"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	Submitter   string    `bson:"submitter" json:"submitter" required:"true"`

	// CorporationManagers is the managers of corporation who can manage the employee
	CorporationManagers []corporationManagerDoc `bson:"corp_managers" json:"-"`

	Md5sumOfOrgSignature string `bson:"md5sum" json:"md5sum"`
	OrgSignature         []byte `bson:"org_signature" json:"-"`
}

func (this *client) CreateOrgCLA(info dbmodels.OrgCLA) (string, error) {
	orgCLA := OrgCLA{
		Platform:    info.Platform,
		OrgID:       info.OrgID,
		RepoID:      dbValueOfRepo(info.OrgID, info.RepoID),
		CLAID:       info.CLAID,
		CLALanguage: info.CLALanguage,
		ApplyTo:     info.ApplyTo,
		OrgEmail:    info.OrgEmail,
		Enabled:     info.Enabled,
		Submitter:   info.Submitter,
	}
	body, err := structToMap(orgCLA)
	if err != nil {
		return "", err
	}

	filterOfDoc, _ := filterOfOrgRepo(info.Platform, info.OrgID, info.RepoID)
	filterOfDoc["cla_language"] = info.CLALanguage
	filterOfDoc["apply_to"] = info.ApplyTo
	filterOfDoc["enabled"] = true

	orgCLAID := ""

	f := func(ctx context.Context) error {
		s, err := this.newDocIfNotExist(ctx, this.orgCLACollection, filterOfDoc, body)
		if err != nil {
			return err
		}
		orgCLAID = s
		return nil
	}

	if err = withContext(f); err != nil {
		return "", err
	}
	return orgCLAID, nil
}

func (this *client) DeleteOrgCLA(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return this.updateDoc(ctx, this.orgCLACollection, filterOfDocID(oid), bson.M{"enabled": false})
	}

	return withContext(f)
}

func (this *client) GetOrgCLA(uid string) (dbmodels.OrgCLA, error) {
	var r dbmodels.OrgCLA

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var v OrgCLA

	f := func(ctx context.Context) error {
		return this.getDoc(ctx, this.orgCLACollection, filterOfDocID(oid), projectOfClaOrg(), &v)
	}

	if err := withContext(f); err != nil {
		return r, err
	}

	return toModelOrgCLA(v), nil
}

func (this *client) ListOrgs(platform string, orgs []string) ([]dbmodels.OrgCLA, error) {
	filter := bson.M{
		"platform": platform,
		"org_id":   bson.M{"$in": orgs},
	}
	filterForClaOrgDoc(filter)

	var v []OrgCLA

	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.orgCLACollection, filter, projectOfClaOrg(), &v)
	}

	if err := withContext(f); err != nil {
		return nil, err
	}

	n := len(v)
	r := make([]dbmodels.OrgCLA, 0, n)
	for _, item := range v {
		r = append(r, toModelOrgCLA(item))
	}

	return r, nil
}

func (this *client) ListOrgCLA(opt dbmodels.OrgCLAListOption) ([]dbmodels.OrgCLA, error) {
	filter, err := filterOfOrgRepo(opt.Platform, opt.OrgID, opt.RepoID)
	if err != nil {
		return nil, err
	}
	if opt.ApplyTo != "" {
		filter["apply_to"] = opt.ApplyTo
	}
	filterForClaOrgDoc(filter)

	var v []OrgCLA

	f := func(ctx context.Context) error {
		return this.getDocs(ctx, this.orgCLACollection, filter, projectOfClaOrg(), &v)
	}

	if err = withContext(f); err != nil {
		return nil, err
	}

	n := len(v)
	r := make([]dbmodels.OrgCLA, 0, n)
	for _, item := range v {
		r = append(r, toModelOrgCLA(item))
	}

	return r, nil
}

func toModelOrgCLA(item OrgCLA) dbmodels.OrgCLA {
	return dbmodels.OrgCLA{
		ID:                   objectIDToUID(item.ID),
		Platform:             item.Platform,
		OrgID:                item.OrgID,
		RepoID:               toNormalRepo(item.RepoID),
		OrgAlias:             item.OrgAlias,
		CLAID:                item.CLAID,
		CLALanguage:          item.CLALanguage,
		ApplyTo:              item.ApplyTo,
		OrgEmail:             item.OrgEmail,
		Enabled:              item.Enabled,
		Submitter:            item.Submitter,
		OrgSignatureUploaded: item.Md5sumOfOrgSignature != "",
	}
}

func projectOfClaOrg() bson.M {
	return bson.M{
		fieldIndividuals:  0,
		fieldEmployees:    0,
		fieldCorporations: 0,
		fieldCorpManagers: 0,
		fieldOrgSignature: 0,
	}
}
