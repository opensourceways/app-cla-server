package mongodb

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldOrgIdentity     = "org_identity"
	fieldCorpManagers    = "corp_managers"
	fieldCorpID          = "corp_id"
	fieldOrgSignature    = "org_signature"
	fieldOrgSignatureTag = "md5sum"
	fieldRepo            = "repo_id"
	fieldOrgAlias        = "org_alias"
	fieldOrgEmail        = "org_email"
	fieldIndividualCLAs  = "individual_clas"
	fieldCorpCLAs        = "corp_clas"
	fieldToken           = "token"
	fieldLinkStatus      = "link_status"
	fieldCLALanguage     = "cla_lang"
	fieldSignings        = "signings"

	// 'enabled' means the doc is used to record the signing data currently.
	// 'unabled' means the doc is invalid, probably because the cla had been changed.
	// 'old' means the signing data is valid. It is changed from enabled because the storage is full.
	linkStatusEnabled = "enabled"
	linkStatusUnabled = "unabled"
	linkStatusDeleted = "deleted"
)

type DOrgRepo struct {
	Platform string `bson:"platform" json:"platform" required:"true"`
	OrgID    string `bson:"org_id" json:"org_id" required:"true"`
	RepoID   string `bson:"repo_id" json:"repo_id"`
}

type cOrgCLA struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	OrgIdentity string `bson:"org_identity" json:"org_identity"`

	DOrgRepo `bson:",inline"`
	OrgAlias string `bson:"org_alias" json:"org_alias"`

	OrgEmail dOrgEmail `bson:"org_email" json:"-"`

	Submitter  string `bson:"submitter" json:"submitter" required:"true"`
	LinkStatus string `bson:"link_status" json:"link_status"`

	IndividualCLAs []dCLA `bson:"individual_clas" json:"-"`
	CorpCLAs       []dCLA `bson:"corp_clas" json:"-"`
}

type dCLA struct {
	URL      string   `bson:"url" json:"url" required:"true"`
	Text     []byte   `bson:"text" json:"text" required:"true"`
	Language string   `bson:"cla_lang" json:"cla_lang" required:"true"`
	Fields   []dField `bson:"fields" json:"fields,omitempty"`

	Md5sumOfOrgSignature string `bson:"md5sum" json:"md5sum,omitempty"`
	OrgSignature         []byte `bson:"org_signature" json:"-"`
}

type dField struct {
	ID          string `bson:"id" json:"id" required:"true"`
	Title       string `bson:"title" json:"title" required:"true"`
	Type        string `bson:"type" json:"type" required:"true"`
	Description string `bson:"description" json:"description,omitempty"`
	Required    bool   `bson:"required" json:"required"`
}

// can't store the email for each orgs, because there is not refreshToken
// if re-authorized to a same email.
type dOrgEmail struct {
	Email    string `bson:"email" json:"email" required:"true"`
	Platform string `bson:"platform" json:"platform" required:"true"`
	Token    []byte `bson:"token" json:"-"`
}

type cIndividualSigning struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	OrgIdentity string `bson:"org_identity" json:"org_identity"`
	LinkStatus  string `bson:"link_status" json:"link_status"`

	// DOrgRepo `bson:",inline"`

	Signings []dIndividualSigning `bson:"signings" json:"-"`
}

type dIndividualSigning struct {
	CLALanguage string `bson:"cla_lang" json:"cla_lang" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`

	Name    string `bson:"name" json:"name" required:"true"`
	Email   string `bson:"email" json:"email" required:"true"`
	Date    string `bson:"date" json:"date" required:"true"`
	Enabled bool   `bson:"enabled" json:"enabled"`

	SigningInfo dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`
}

type cCorpSigning struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	OrgIdentity string `bson:"org_identity" json:"org_identity"`
	LinkStatus  string `bson:"link_status" json:"link_status"`

	Signings []dCorpSigning `bson:"signings" json:"-"`
}

type dCorpSigning struct {
	CLALanguage string `bson:"cla_lang" json:"cla_lang" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`

	CorporationName string `bson:"corp_name" json:"corp_name" required:"true"`
	AdminEmail      string `bson:"admin_email" json:"admin_email" required:"true"`
	AdminName       string `bson:"admin_name" json:"admin_name" required:"true"`
	Date            string `bson:"date" json:"date" required:"true"`

	SigningInfo dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`

	PDFUploaded bool   `bson:"pdf_uploaded" json:"pdf_uploaded"`
	PDF         []byte `bson:"pdf" json:"pdf,omitempty"`
}

type cCorpManager struct {
	ID primitive.ObjectID `bson:"_id" json:"-"`

	OrgIdentity string `bson:"org_identity" json:"org_identity"`
	LinkStatus  string `bson:"link_status" json:"link_status"`

	CorpManagers []dCorpManager `bson:"corp_managers" json:"-"`
}

type dCorpManager struct {
	ID               string `bson:"id" json:"id" required:"true"`
	Name             string `bson:"name" json:"name" required:"true"`
	Role             string `bson:"role" json:"role" required:"true"`
	Email            string `bson:"email"  json:"email" required:"true"`
	CorpID           string `bson:"corp_id" json:"corp_id" required:"true"`
	Password         string `bson:"password" json:"password" required:"true"`
	InitialPWChanged bool   `bson:"changed" json:"changed"`
}

func orgIdentity(v *dbmodels.OrgRepo) string {
	return genOrgIdentity(v.Platform, v.OrgID, v.RepoID)
}

func genOrgIdentity(platform, org, repo string) string {
	if repo == "" {
		return fmt.Sprintf("%s/%s", platform, org)
	}
	return fmt.Sprintf("%s/%s/%s", platform, org, repo)
}

func parseOrgIdentity(identity string) *dbmodels.OrgRepo {
	r := dbmodels.OrgRepo{}

	v := strings.Split(identity, "/")
	switch len(v) {
	case 2:
		r.Platform = v[0]
		r.OrgID = v[1]
		r.RepoID = ""
	case 3:
		r.Platform = v[0]
		r.OrgID = v[1]
		r.RepoID = v[2]
	}

	r.Platform = identity
	r.OrgID = ""
	r.RepoID = ""

	return &r
}

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}
