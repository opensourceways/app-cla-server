package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const fieldDeletedAt = "deleted_at"

func toIndividualSigningDO(is *domain.IndividualSigning) individualSigningDO {
	var logs []IndividualSigningLogDO
	for _, v := range is.Logs {
		logs = append(logs, toIndividualLogDO(v))
	}

	return individualSigningDO{
		CLAId:    is.Link.CLAId,
		LinkId:   is.Link.Id,
		Language: is.Link.Language.Language(),
		Date:     is.Date,
		AllInfo:  is.AllInfo,
		RepDO:    toRepDO(&is.Rep),
		Domain:   is.Rep.EmailAddr.Domain(),
		Logs:     logs,
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
	Logs     []IndividualSigningLogDO `bson:"logs"     json:"logs"`

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

type IndividualSigningLogDO struct {
	Date   string `bson:"date" json:"time" required:"true"`
	ClaId  string `bson:"cla_id" json:"cla_id" required:"true"`
	Action string `bson:"action" json:"action" required:"true"`
}

func toIndividualLogDO(log domain.IndividualSigningLog) IndividualSigningLogDO {
	return IndividualSigningLogDO{
		Date:   log.Date,
		ClaId:  log.ClaId,
		Action: log.Action,
	}
}

func (do *IndividualSigningLogDO) toIndividualSigningLog() domain.IndividualSigningLog {
	return domain.IndividualSigningLog{
		Date:   do.Date,
		ClaId:  do.ClaId,
		Action: do.Action,
	}
}

func (do *individualSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
