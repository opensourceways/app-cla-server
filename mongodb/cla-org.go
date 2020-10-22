package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	claOrgCollection     = "cla_orgs"
	fieldIndividuals     = "individuals"
	fieldEmployees       = "employees"
	fieldCorporations    = "corporations"
	fieldCorpoManagers   = "corp_managers"
	fieldCorporationID   = "corp_id"
	fieldOrgSignature    = "org_signature"
	fieldOrgSignatureTag = "has_org_signature"
	fieldRepo            = "repo_id"
)

func filterForClaOrgDoc(filter bson.M) {
	filter["enabled"] = true
}

type CLAOrg struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	CreatedAt   time.Time `bson:"created_at" json:"-"`
	UpdatedAt   time.Time `bson:"updated_at" json:"-"`
	Platform    string    `bson:"platform" json:"platform" required:"true"`
	OrgID       string    `bson:"org_id" json:"org_id" required:"true"`
	RepoID      string    `bson:"repo_id" json:"repo_id"`
	CLAID       string    `bson:"cla_id" json:"cla_id" required:"true"`
	CLALanguage string    `bson:"cla_language" json:"cla_language" required:"true"`
	ApplyTo     string    `bson:"apply_to" json:"apply_to" required:"true"`
	OrgEmail    string    `bson:"org_email" json:"org_email" required:"true"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	Submitter   string    `bson:"submitter" json:"submitter" required:"true"`

	// Individuals is the cla signing information of ordinary contributors
	// key is the email of contributor
	Individuals []individualSigningDoc `bson:"individuals" json:"-"`

	// Corporations is the cla signing information of corporation
	// key is the email suffix of corporation
	Corporations []corporationSigningDoc `bson:"corporations" json:"-"`

	// CorporationManagers is the managers of corporation who can manage the employee
	CorporationManagers []corporationManagerDoc `bson:"corp_managers" json:"-"`

	HasOrgSignature bool   `bson:"has_org_signature" json:"has_org_signature"`
	OrgSignature    []byte `bson:"org_signature" json:"-"`
}

func (c *client) CreateBindingBetweenCLAAndOrg(info dbmodels.CLAOrg) (string, error) {
	claOrg := CLAOrg{
		Platform:        info.Platform,
		OrgID:           info.OrgID,
		RepoID:          dbValueOfRepo(info.OrgID, info.RepoID),
		CLAID:           info.CLAID,
		CLALanguage:     info.CLALanguage,
		ApplyTo:         info.ApplyTo,
		OrgEmail:        info.OrgEmail,
		Enabled:         info.Enabled,
		Submitter:       info.Submitter,
		HasOrgSignature: info.OrgSignatureUploaded,
	}
	body, err := structToMap(claOrg)
	if err != nil {
		return "", err
	}

	filterOfDoc, _ := filterOfOrgRepo(info.Platform, info.OrgID, info.RepoID)
	filterOfDoc["cla_language"] = info.CLALanguage
	filterOfDoc["apply_to"] = info.ApplyTo
	filterOfDoc["enabled"] = true

	claOrgID := ""

	f := func(ctx context.Context) error {
		s, err := c.newDocIfNotExist(ctx, claOrgCollection, filterOfDoc, body)
		if err != nil {
			return err
		}
		claOrgID = s
		return nil
	}

	if err = withContext(f); err != nil {
		return "", err
	}
	return claOrgID, nil
}

func (c *client) DeleteBindingBetweenCLAAndOrg(uid string) error {
	oid, err := toObjectID(uid)
	if err != nil {
		return err
	}

	f := func(ctx context.Context) error {
		return c.updateDoc(ctx, claOrgCollection, filterOfDocID(oid), bson.M{"enabled": false})
	}

	return withContext(f)
}

func (c *client) GetBindingBetweenCLAAndOrg(uid string) (dbmodels.CLAOrg, error) {
	var r dbmodels.CLAOrg

	oid, err := toObjectID(uid)
	if err != nil {
		return r, err
	}

	var v CLAOrg

	f := func(ctx context.Context) error {
		return c.getDoc(ctx, claOrgCollection, filterOfDocID(oid), projectOfClaOrg(), &v)
	}

	if err := withContext(f); err != nil {
		return r, err
	}

	return toModelCLAOrg(v), nil
}

func (c *client) ListBindingBetweenCLAAndOrg(opt dbmodels.CLAOrgListOption) ([]dbmodels.CLAOrg, error) {
	if (opt.RepoID != "" && len(opt.OrgID) > 0) || (opt.RepoID == "" && len(opt.OrgID) == 0) {
		return nil, fmt.Errorf("need specify multiple orgs or a single repo")
	}

	info := struct {
		Platform string `json:"platform" required:"true"`
		RepoID   string `json:"repo_id,omitempty"`
		ApplyTo  string `json:"apply_to,omitempty"`
	}{
		Platform: opt.Platform,
		RepoID:   opt.RepoID,
		ApplyTo:  opt.ApplyTo,
	}

	filter, err := structToMap(info)
	if err != nil {
		return nil, err
	}
	filterForClaOrgDoc(filter)
	if len(opt.OrgID) > 0 {
		filter["org_id"] = bson.M{"$in": opt.OrgID}
	}

	var v []CLAOrg

	f := func(ctx context.Context) error {
		return c.getDocs(ctx, claOrgCollection, filter, projectOfClaOrg(), &v)
	}

	if err = withContext(f); err != nil {
		return nil, err
	}

	n := len(v)
	r := make([]dbmodels.CLAOrg, 0, n)
	for _, item := range v {
		r = append(r, toModelCLAOrg(item))
	}

	return r, nil
}

func toModelCLAOrg(item CLAOrg) dbmodels.CLAOrg {
	return dbmodels.CLAOrg{
		ID:                   objectIDToUID(item.ID),
		Platform:             item.Platform,
		OrgID:                item.OrgID,
		RepoID:               toNormalRepo(item.RepoID),
		CLAID:                item.CLAID,
		CLALanguage:          item.CLALanguage,
		ApplyTo:              item.ApplyTo,
		OrgEmail:             item.OrgEmail,
		Enabled:              item.Enabled,
		Submitter:            item.Submitter,
		OrgSignatureUploaded: item.HasOrgSignature,
	}
}

func projectOfClaOrg() bson.M {
	return bson.M{
		fieldIndividuals:   0,
		fieldEmployees:     0,
		fieldCorporations:  0,
		fieldCorpoManagers: 0,
		fieldOrgSignature:  0,
	}
}
