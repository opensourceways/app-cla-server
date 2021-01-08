package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldLinkID         = "link_id"
	fieldLinkStatus     = "link_status"
	fieldCorpID         = "corp_id"
	fieldSignings       = "signings"
	fieldCLALang        = "cla_lang"
	fieldOrgEmail       = "org_email"
	fieldOrgAlias       = "org_alias"
	fieldOrgIdentity    = "org_identity"
	fieldIndividualCLAs = "individual_clas"
	fieldCorpCLAs       = "corp_clas"

	// 'ready' means the doc is ready to record the signing data currently.
	// 'deleted' means the signing data is invalid.
	linkStatusReady   = "ready"
	linkStatusDeleted = "deleted"
)

type dCorpSigningPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}

type cVerificationCode struct {
	Email   string `bson:"email" json:"email" required:"true"`
	Code    string `bson:"code" json:"code" required:"true"`
	Purpose string `bson:"purpose" json:"purpose" required:"true"`
	Expiry  int64  `bson:"expiry" json:"expiry" required:"true"`
}

type cIndividualSigning struct {
	LinkID     string `bson:"link_id" json:"link_id" required:"true"`
	LinkStatus string `bson:"link_status" json:"link_status" required:"true"`

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
	LinkID     string `bson:"link_id" json:"link_id" required:"true"`
	LinkStatus string `bson:"link_status" json:"link_status" required:"true"`

	OrgIdentity string `bson:"org_identity" json:"org_identity" required:"true"`
	OrgEmail    string `bson:"org_email" json:"org_email" required:"true"`
	OrgAlias    string `bson:"org_alias" json:"org_alias" required:"true"`

	Signings []dCorpSigning `bson:"signings" json:"-"`
	Managers []dCorpManager `bson:"corp_managers" json:"-"`
}

type dCorpSigning struct {
	CLALanguage string `bson:"cla_lang" json:"cla_lang" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`

	CorporationName string `bson:"corp_name" json:"corp_name" required:"true"`
	AdminEmail      string `bson:"admin_email" json:"admin_email" required:"true"`
	AdminName       string `bson:"admin_name" json:"admin_name" required:"true"`
	Date            string `bson:"date" json:"date" required:"true"`

	SigningInfo dbmodels.TypeSigningInfo `bson:"info" json:"info,omitempty"`
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

type cOrgEmail struct {
	Email    string `bson:"email" json:"email" required:"true"`
	Platform string `bson:"platform" json:"platform" required:"true"`
	Token    []byte `bson:"token" json:"-"`
}

type DCLAInfo struct {
	Fields           []dField `bson:"fields" json:"fields,omitempty"`
	Language         string   `bson:"cla_lang" json:"cla_lang" required:"true"`
	CLAHash          string   `bson:"cla_hash" json:"cla_hash" required:"true"`
	OrgSignatureHash string   `bson:"org_signature_hash" json:"org_signature_hash,omitempty"`
}

type cLink struct {
	LinkID     string `bson:"link_id" json:"link_id" required:"true"`
	LinkStatus string `bson:"link_status" json:"link_status"`

	Platform  string `bson:"platform" json:"platform" required:"true"`
	OrgID     string `bson:"org_id" json:"org_id" required:"true"`
	RepoID    string `bson:"repo_id" json:"repo_id"`
	OrgAlias  string `bson:"org_alias" json:"org_alias"`
	Submitter string `bson:"submitter" json:"submitter" required:"true"`

	OrgEmail cOrgEmail `bson:"org_email" json:"org_email" required:"true"`

	IndividualCLAs []dCLA `bson:"individual_clas" json:"-"`
	CorpCLAs       []dCLA `bson:"corp_clas" json:"-"`
}

type dCLA struct {
	URL          string `bson:"url" json:"url" required:"true"`
	Text         string `bson:"text" json:"text" required:"true"`
	OrgSignature []byte `bson:"org_signature" json:"-"`

	DCLAInfo `bson:",inline"`
}

type dField struct {
	ID          string `bson:"id" json:"id" required:"true"`
	Title       string `bson:"title" json:"title" required:"true"`
	Type        string `bson:"type" json:"type" required:"true"`
	Description string `bson:"description" json:"description,omitempty"`
	Required    bool   `bson:"required" json:"required"`
}

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}
