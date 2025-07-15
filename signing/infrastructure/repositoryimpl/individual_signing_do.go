package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const fieldDeletedAt = "deleted_at"

func toIndividualSigningDO(is *domain.IndividualSigning) individualSigningDO {
	return individualSigningDO{
		CLAId:                   is.Link.CLAId,
		LinkId:                  is.Link.Id,
		Language:                is.Link.Language.Language(),
		Date:                    is.Date,
		AllInfo:                 is.AllInfo,
		RepDO:                   toRepDO(&is.Rep),
		Domain:                  is.Rep.EmailAddr.Domain(),
		IndividualSigningLogsDO: toIndividualSigningLogsDO(is.Logs),
	}
}

// individualSigningDO
type individualSigningDO struct {
	CLAId                   string `bson:"cla_id"      json:"cla_id"   required:"true"`
	LinkId                  string `bson:"link_id"     json:"link_id"  required:"true"`
	Language                string `bson:"lang"        json:"lang"     required:"true"`
	Date                    string `bson:"date"        json:"date"     required:"true"`
	AllInfo                 anyDoc `bson:"info"        json:"info,omitempty"`
	RepDO                   `bson:",inline"`
	IndividualSigningLogsDO `bson:",inline"`

	Domain    string `bson:"domain"      json:"domain"  required:"true"`
	Version   int    `bson:"version"     json:"-"`
	Deleted   bool   `bson:"deleted"     json:"deleted"`
	DeletedAt int64  `bson:"deleted_at"  json:"deleted_at,omitempty"`
}

func (do *individualSigningDO) toIndividualSigning() domain.IndividualSigning {
	var logs []domain.IndividualSigningLog
	for _, v := range do.Logs {
		logs = append(logs, v.toIndividualSigningLog())
	}

	return domain.IndividualSigning{
		Link: domain.LinkInfo{
			Id: do.LinkId,
			CLAInfo: domain.CLAInfo{
				CLAId:    do.CLAId,
				Language: dp.CreateLanguage(do.Language),
			},
		},
		Rep:     do.toRep(),
		Date:    do.Date,
		AllInfo: do.AllInfo,
		Version: do.Version,
		Logs:    logs,
	}
}

type IndividualSigningLogsDO struct {
	Logs []individualSigningLogDO `bson:"logs"     json:"logs"`
}

func toIndividualSigningLogsDO(logs []domain.IndividualSigningLog) IndividualSigningLogsDO {
	var dos []individualSigningLogDO
	for _, v := range logs {
		dos = append(dos, toIndividualSigningLogDO(v))
	}

	return IndividualSigningLogsDO{dos}
}

type individualSigningLogDO struct {
	Date   string `bson:"date"   json:"date"   required:"true"`
	ClaId  string `bson:"cla_id" json:"cla_id" required:"true"`
	Action string `bson:"action" json:"action" required:"true"`
}

func toIndividualSigningLogDO(log domain.IndividualSigningLog) individualSigningLogDO {
	return individualSigningLogDO{
		Date:   log.Date,
		ClaId:  log.ClaId,
		Action: log.Action,
	}
}

func (do *individualSigningLogDO) toIndividualSigningLog() domain.IndividualSigningLog {
	return domain.IndividualSigningLog{
		Date:   do.Date,
		ClaId:  do.ClaId,
		Action: do.Action,
	}
}

func (do *individualSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
