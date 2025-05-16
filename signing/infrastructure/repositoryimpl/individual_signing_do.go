package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const fieldDeletedAt = "deleted_at"

func toIndividualSigningDO(is *domain.IndividualSigning) individualSigningDO {
	return individualSigningDO{
		CLAId:            is.Link.CLAId,
		LinkId:           is.Link.Id,
		Language:         is.Link.Language.Language(),
		Date:             is.Date,
		AllInfo:          is.AllInfo,
		RepDO:            toRepDO(&is.Rep),
		Domain:           is.Rep.EmailAddr.Domain(),
		AgreementVersion: is.Link.AgreementVersion,
	}
}

// individualSigningDO
type individualSigningDO struct {
	CLAId    string `bson:"cla_id"      json:"cla_id"   required:"true"`
	LinkId   string `bson:"link_id"     json:"link_id"  required:"true"`
	Language string `bson:"lang"        json:"lang"     required:"true"`
	Date     string `bson:"date"        json:"date"     required:"true"`
	AllInfo  anyDoc `bson:"info"        json:"info,omitempty"`
	RepDO    `bson:",inline"`

	Domain           string `bson:"domain"      json:"domain"  required:"true"`
	Deleted          bool   `bson:"deleted"     json:"deleted"`
	DeletedAt        int64  `bson:"deleted_at"  json:"deleted_at,omitempty"`
	AgreementVersion int    `bson:"agreement_version" json:"agreement_version"`
}

func (do *individualSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *individualSigningDO) toIndividualSigning() domain.IndividualSigning {
	return domain.IndividualSigning{
		Link: domain.LinkInfo{
			Id: do.LinkId,
			CLAInfo: domain.CLAInfo{
				CLAId:            do.CLAId,
				Language:         dp.CreateLanguage(do.Language),
				AgreementVersion: do.AgreementVersion,
			},
		},
		Rep:     do.toRep(),
		Date:    do.Date,
		AllInfo: do.AllInfo,
	}
}
