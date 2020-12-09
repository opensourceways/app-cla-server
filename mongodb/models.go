package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldCorpID      = "corp_id"
	fieldLinkStatus  = "link_status"
	fieldOrgIdentity = "org_identity"
	fieldLinkID      = "link_id"
	fieldSignings    = "signings"
	fieldCLALang     = "cla_lang"

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
	Language string   `bson:"cla_lang" json:"cla_lang" required:"true"`
	Fields   []dField `bson:"fields" json:"fields,omitempty"`
	CLAHash  string   `bson:"cla_hash" json:"cla_hash" required:"true"`
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
	OrgIdentity string `bson:"org_identity" json:"org_identity"`
	LinkStatus  string `bson:"link_status" json:"link_status"`

	DCLAInfo `bson:",inline"`

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

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}
