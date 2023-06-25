package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func toEmployeeSigningDO(es *domain.EmployeeSigning) employeeSigningDO {
	return employeeSigningDO{
		Id:       es.Id,
		CLAId:    es.CLA.CLAId,
		Language: es.CLA.Language.Language(),
		Name:     es.Rep.Name.Name(),
		Email:    es.Rep.EmailAddr.EmailAddr(),
		Date:     es.Date,
		Enabled:  es.Enabled,
		Deleted:  es.Deleted,
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

// employeeSigningDO
type employeeSigningDO struct {
	Id       string                 `bson:"id"       json:"id"       required:"true"`
	CLAId    string                 `bson:"cla_id"   json:"cla_id"   required:"true"`
	Language string                 `bson:"lang"     json:"lang"     required:"true"`
	Name     string                 `bson:"name"     json:"name"     required:"true"`
	Email    string                 `bson:"email"    json:"email"    required:"true"`
	Date     string                 `bson:"date"     json:"date"     required:"true"`
	Enabled  bool                   `bson:"enabled"  json:"enabled"`
	Deleted  bool                   `bson:"deleted"  json:"deleted,omitempty"`
	AllInfo  anyDoc                 `bson:"info"     json:"info,omitempty"`
	Logs     []employeeSigningLogDO `bson:"logs"     json:"logs"`
}

func (do *employeeSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
