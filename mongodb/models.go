package mongodb

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	fieldLinkID     = "link_id"
	fieldLinkStatus = "link_status"
	fieldCorpID     = "corp_id"
	fieldSignings   = "signings"
	fieldCLALang    = "cla_lang"

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

func memberNameOfSignings(key string) string {
	return fmt.Sprintf("%s.%s", fieldSignings, key)
}
