package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const fieldEnabled = "enabled"

func toEmployeeSigningDO(es *domain.EmployeeSigning) employeeSigningDO {
	return employeeSigningDO{
		Id:       es.Id,
		CLAId:    es.CLA.CLAId,
		Language: es.CLA.Language.Language(),
		RepDO:    toRepDO(&es.Rep),
		Date:     es.Date,
		Enabled:  es.Enabled,
		AllInfo:  es.AllInfo,
		Logs:     toEmployeeSigningLogDOs(es.Logs),
	}
}

func toEmployeeSigningLogDOs(logs []domain.EmployeeSigningLog) []employeeSigningLogDO {
	r := make([]employeeSigningLogDO, len(logs))

	for i := range logs {
		item := &logs[i]

		r[i] = employeeSigningLogDO{
			Time:   item.Time,
			Action: item.Action,
		}
	}

	return r
}

// employeeSigningLogDO
type employeeSigningLogDO struct {
	Time   int64  `bson:"time"     json:"time"     required:"true"`
	Action string `bson:"action"     json:"action"     required:"true"`
}

func (do *employeeSigningLogDO) toEmployeeSigningLog() domain.EmployeeSigningLog {
	return domain.EmployeeSigningLog{
		Time:   do.Time,
		Action: do.Action,
	}
}

// employeeSigningDO
type employeeSigningDO struct {
	Id       string                 `bson:"id"       json:"id"       required:"true"`
	CLAId    string                 `bson:"cla_id"   json:"cla_id"   required:"true"`
	Language string                 `bson:"lang"     json:"lang"     required:"true"`
	Date     string                 `bson:"date"     json:"date"     required:"true"`
	Enabled  bool                   `bson:"enabled"  json:"enabled"`
	AllInfo  anyDoc                 `bson:"info"     json:"info,omitempty"`
	Logs     []employeeSigningLogDO `bson:"logs"     json:"logs"`

	RepDO `bson:",inline"`
}

func (do *employeeSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *employeeSigningDO) toEmployeeSigning(es *domain.EmployeeSigning) {
	*es = domain.EmployeeSigning{
		Id:      do.Id,
		Rep:     do.RepDO.toRep(),
		Date:    do.Date,
		Enabled: do.Enabled,
		AllInfo: do.AllInfo,
		Logs:    do.toEmployeeSigningLogs(),
	}

	es.CLA.CLAId = do.CLAId
	es.CLA.Language = dp.CreateLanguage(do.Language)
}

func (do *employeeSigningDO) toEmployeeSigningLogs() []domain.EmployeeSigningLog {
	r := make([]domain.EmployeeSigningLog, len(do.Logs))

	for i := range do.Logs {
		r[i] = do.Logs[i].toEmployeeSigningLog()
	}

	return r
}
