package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldCorpID         = "corp_id"
	fieldLinkStatus     = "link_status"
	fieldOrgIdentity    = "org_identity"
	fieldLinkID         = "link_id"
	fieldSignings       = "signings"
	fieldCLALang        = "cla_lang"
	fieldSingingCLAInfo = "cla_info"

	// 'ready' means the doc is ready to record the signing data currently.
	// 'unready' means the doc is not ready.
	// 'deleted' means the signing data is invalid.
	linkStatusReady   = "ready"
	linkStatusUnready = "unready"
	linkStatusDeleted = "deleted"
)

type dCorpSigningPDF struct {
	LinkID string `bson:"link_id" json:"link_id" required:"true"`
	CorpID string `bson:"corp_id" json:"corp_id" required:"true"`
	PDF    []byte `bson:"pdf" json:"pdf,omitempty"`
}

type DCLAInfo struct {
	Fields           []dField `bson:"fields" json:"fields,omitempty"`
	Language         string   `bson:"cla_lang" json:"cla_lang" required:"true"`
	CLAHash          string   `bson:"cla_hash" json:"cla_hash" required:"true"`
	OrgSignatureHash string   `bson:"org_signature_hash" json:"org_signature_hash"`
}

type dField struct {
	ID          string `bson:"id" json:"id" required:"true"`
	Title       string `bson:"title" json:"title" required:"true"`
	Type        string `bson:"type" json:"type" required:"true"`
	Description string `bson:"description" json:"description,omitempty"`
	Required    bool   `bson:"required" json:"required"`
}

type cIndividualSigning struct {
	LinkID      string `bson:"link_id" json:"link_id" required:"true"`
	LinkStatus  string `bson:"link_status" json:"link_status"`
	OrgIdentity string `bson:"org_identity" json:"org_identity"`

	CLAInfo  []DCLAInfo           `bson:"cla_info"`
	Signings []dIndividualSigning `bson:"signings"`
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
	LinkStatus string `bson:"link_status" json:"link_status"`

	OrgIdentity string `bson:"org_identity" json:"org_identity"`
	OrgEmail    string `bson:"org_email" json:"org_email"`
	OrgAlias    string `bson:"org_alias" json:"org_alias"`

	CLAInfo  []DCLAInfo     `bson:"cla_info"`
	Signings []dCorpSigning `bson:"signings"`
	Managers []dCorpManager `bson:"corp_managers"`
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

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}

func toModelOfCLAFields(fields []dField) []dbmodels.Field {
	fs := make([]dbmodels.Field, 0, len(fields))
	for _, v := range fields {
		fs = append(fs, dbmodels.Field{
			ID:          v.ID,
			Title:       v.Title,
			Type:        v.Type,
			Description: v.Description,
			Required:    v.Required,
		})
	}
	return fs
}
