package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldLinkID         = "link_id"
	fieldLinkStatus     = "link_status"
	fieldPlatform       = "platform"
	fieldOrg            = "org"
	fieldRepo           = "repo"
	fieldCorpID         = "corp_id"
	fieldSignings       = "signings"
	fieldDeleted        = "deleted"
	fieldLang           = "lang"
	fieldOrgEmail       = "org_email"
	fieldOrgAlias       = "org_alias"
	fieldOrgIdentity    = "org_identity"
	fieldIndividualCLAs = "individual_clas"
	fieldCorpCLAs       = "corp_clas"
	fieldCLAInfos       = "cla_infos"
	fieldCorpManagers   = "corp_managers"
	fieldOrgSignature   = "org_signature"
	fieldPassword       = "password"
	fieldChanged        = "changed"
	fieldFields         = "fields"
	fieldCLAHash        = "cla_hash"
	fieldSignatureHash  = "signature_hash"
	fieldEmail          = "email"
	fieldPurpose        = "purpose"
	fieldCode           = "code"
	fieldExpiry         = "expiry"
	fieldToken          = "token"
	fieldRole           = "role"
	fieldName           = "name"
	fieldID             = "id"
	fieldPDF            = "pdf"
	fieldDate           = "date"
	fieldCorp           = "corp"
	fieldEnabled        = "enabled"

	// 'ready' means the doc is ready to record the signing data currently.
	// 'deleted' means the signing data is invalid.
	linkStatusReady   = "ready"
	linkStatusDeleted = "deleted"
)

type dCLAPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	Apply  string `bson:"apply" json:"apply" required:"true"`
	Lang   string `bson:"lang" json:"lang" required:"true"`
	Hash   string `bson:"hash" json:"hash" required:"true"`
	PDF    []byte `bson:"pdf" json:"-"`
}

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

	CLAInfos []DCLAInfo           `bson:"cla_infos" json:"cla_infos,omitempty"`
	Signings []dIndividualSigning `bson:"signings" json:"-"`
}

type dIndividualSigning struct {
	CLALanguage string `bson:"lang" json:"lang" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`

	ID      string `bson:"id" json:"id,omitempty"`
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

	CLAInfos []DCLAInfo     `bson:"cla_infos" json:"cla_infos,omitempty"`
	Signings []dCorpSigning `bson:"signings" json:"-"`
	Managers []dCorpManager `bson:"corp_managers" json:"-"`
	Deleted  []dCorpSigning `bson:"deleted" json:"-"`
}

type dCorpSigning struct {
	CLALanguage string `bson:"lang" json:"lang" required:"true"`
	CorpID      string `bson:"corp_id" json:"corp_id" required:"true"`
	CorpName    string `bson:"corp" json:"corp" required:"true"`

	AdminEmail string `bson:"email" json:"email" required:"true"`
	AdminName  string `bson:"name" json:"name" required:"true"`
	Date       string `bson:"date" json:"date" required:"true"`

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
	Language         string   `bson:"lang" json:"lang" required:"true"`
	CLAHash          string   `bson:"cla_hash" json:"cla_hash" required:"true"`
	OrgSignatureHash string   `bson:"signature_hash" json:"signature_hash,omitempty"`
}

type cLink struct {
	LinkID     string `bson:"link_id" json:"link_id" required:"true"`
	LinkStatus string `bson:"link_status" json:"link_status"`

	Platform  string `bson:"platform" json:"platform" required:"true"`
	Org       string `bson:"org" json:"org" required:"true"`
	Repo      string `bson:"repo" json:"repo"`
	OrgAlias  string `bson:"org_alias" json:"org_alias"`
	Submitter string `bson:"submitter" json:"submitter" required:"true"`

	OrgEmail cOrgEmail `bson:"org_email" json:"-"`

	IndividualCLAs []dCLA `bson:"individual_clas" json:"-"`
	CorpCLAs       []dCLA `bson:"corp_clas" json:"-"`
}

type dCLA struct {
	URL          string `bson:"url" json:"url" required:"true"`
	Text         string `bson:"text" json:"text,omitempty"`
	OrgSignature []byte `bson:"org_signature" json:"-"`

	DCLAInfo `bson:",inline"`
}

type dField struct {
	ID          string `bson:"id" json:"id" required:"true"`
	Title       string `bson:"title" json:"title" required:"true"`
	Type        string `bson:"type" json:"type" required:"true"`
	Description string `bson:"desc" json:"desc,omitempty"`
	Required    bool   `bson:"required" json:"required"`
}

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}
