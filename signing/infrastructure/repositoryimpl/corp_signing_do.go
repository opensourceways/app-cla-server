package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

const (
	fieldRep     = "rep"
	fieldCorp    = "corp"
	fieldName    = "name"
	fieldEmail   = "email"
	fieldDomain  = "domain"
	fieldLinkId  = "link_id"
	fieldVersion = "version"
)

func toCorpSigningDO(v *domain.CorpSigning) corpSigningDO {
	link := &v.Link

	return corpSigningDO{
		Date:        v.Date,
		CLAId:       link.CLAId,
		LinkId:      link.Id,
		CLALanguage: link.Language.Language(),
		Rep:         toRepDO(&v.Rep),
		Corp:        toCorpDO(&v.Corp),
		AllInfo:     v.AllInfo,
	}
}

// corpSigningDO
type corpSigningDO struct {
	Date        string `bson:"date"     json:"date"     required:"true"`
	CLAId       string `bson:"cla_id"   json:"cla_id"   required:"true"`
	LinkId      string `bson:"link_id"  json:"link_id"  required:"true"`
	CLALanguage string `bson:"lang"     json:"lang"     required:"true"`
	Rep         repDO  `bson:"rep"      json:"rep"      required:"true"`
	Corp        corpDO `bson:"corp"     json:"corp"     required:"true"`
	AllInfo     anyDoc `bson:"info"     json:"info,omitempty"`
	Version     int    `bson:"version"  json:"-"`
}

func (do *corpSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

// representative DO
type repDO struct {
	Name  string `bson:"name"  json:"name"  required:"true"`
	Email string `bson:"email" json:"email" required:"true"`
}

func toRepDO(v *domain.Representative) repDO {
	return repDO{
		Name:  v.Name.Name(),
		Email: v.EmailAddr.EmailAddr(),
	}
}

// corporation DO
type corpDO struct {
	Name    string   `bson:"name"     json:"name"      required:"true"`
	Domain  string   `bson:"domain"   json:"domain"    required:"true"`
	Domains []string `bson:"domains"  json:"domains"   required:"true"`
}

func toCorpDO(v *domain.Corporation) corpDO {
	return corpDO{
		Name:    v.Name.CorpName(),
		Domain:  v.PrimaryEmailDomain,
		Domains: v.AllEmailDomains,
	}
}
