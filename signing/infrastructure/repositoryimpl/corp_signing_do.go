package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

const (
	fieldPDF            = "pdf"
	fieldRep            = "rep"
	fieldType           = "type"
	fieldDate           = "date"
	fieldCorp           = "corp"
	fieldName           = "name"
	fieldLang           = "lang"
	fieldAdmin          = "admin"
	fieldEmail          = "email"
	fieldHasPDF         = "has_pdf"
	fieldLinkId         = "link_id"
	fieldDomain         = "domain"
	fieldDomains        = "domains"
	fieldDeleted        = "deleted"
	fieldVersion        = "version"
	fieldManagers       = "managers"
	fieldEmployees      = "employees"
	fieldTriggered      = "triggered"
	fieldCLANotify      = "cla_notify"
	fieldPendingCLAId   = "pending_cla_id"
	fieldCLANotifyCount = "cla_notify_count"
	fieldCLANotifyTime  = "cla_notify_time"
)

func toCorpSigningDO(v *domain.CorpSigning) corpSigningDO {
	link := &v.Link

	return corpSigningDO{
		Date:              v.Date,
		CLAId:             link.CLAId,
		LinkId:            link.Id,
		Language:          link.Language.Language(),
		Rep:               toRepDO(&v.Rep),
		Corp:              toCorpDO(&v.Corp),
		AllInfo:           v.AllInfo,
		CorpSigningLogsDO: toCorpSigningLogsDO(v.Logs),
	}
}

func toCorpSigningDOForMigrate(v *domain.CorpSigning) corpSigningDO {
	link := &v.Link

	return corpSigningDO{
		Date:              v.Date,
		CLAId:             link.CLAId,
		LinkId:            link.Id,
		Language:          link.Language.Language(),
		Rep:               toRepDO(&v.Rep),
		Corp:              toCorpDO(&v.Corp),
		AllInfo:           v.AllInfo,
		Admin:             toManagerDO(&v.Admin),
		Managers:          toManagerDOs(v.Managers),
		Employees:         toEmployeeSigningDOs(v.Employees),
		CorpSigningLogsDO: toCorpSigningLogsDO(v.Logs),
	}
}

// corpSigningDO
type corpSigningDO struct {
	Id       primitive.ObjectID `bson:"_id"      json:"-"`
	Date     string             `bson:"date"     json:"date"     required:"true"`
	CLAId    string             `bson:"cla_id"   json:"cla_id"   required:"true"`
	LinkId   string             `bson:"link_id"  json:"link_id"  required:"true"`
	Language string             `bson:"lang"     json:"lang"     required:"true"`
	Rep      RepDO              `bson:"rep"      json:"rep"      required:"true"`
	Corp     corpDO             `bson:"corp"     json:"corp"     required:"true"`
	AllInfo  anyDoc             `bson:"info"     json:"info,omitempty"`

	PDF       []byte              `bson:"pdf"           json:"pdf,omitempty"`
	HasPDF    bool                `bson:"has_pdf"       json:"has_pdf"`
	Admin     managerDO           `bson:"admin"         json:"admin"`
	Managers  []managerDO         `bson:"managers"      json:"managers"`
	Employees []employeeSigningDO `bson:"employees"     json:"employees"`
	Deleted   []employeeSigningDO `bson:"deleted"       json:"deleted"`
	Version   int                 `bson:"version"       json:"-"`
	ClaNotify string              `bson:"cla_notify"    json:"cla_notify"`

	PendingCLAId   string `bson:"pending_cla_id"    json:"pending_cla_id"`
	ClaNotifyCount int    `bson:"cla_notify_count"  json:"cla_notify_count"`
	ClaNotifyTime  int64  `bson:"cla_notify_time"   json:"cla_notify_time"`

	CorpSigningLogsDO `bson:",inline"`

	Triggered bool `bson:"triggered" json:"triggered,omitempty"`
}

func (do *corpSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *corpSigningDO) index() string {
	return do.Id.Hex()
}

func (do *corpSigningDO) toCorpSigningSummary() repository.CorpSigningSummary {
	return repository.CorpSigningSummary{
		Id:   do.index(),
		Rep:  do.Rep.toRep(),
		Date: do.Date,
		Corp: do.Corp.toCorp(),
		Link: domain.LinkInfo{
			Id: do.LinkId,
			CLAInfo: domain.CLAInfo{
				CLAId:    do.CLAId,
				Language: dp.CreateLanguage(do.Language),
			},
		},
		Admin:          do.Admin.toManager(),
		HasPDF:         do.HasPDF,
		CLANotify:      do.ClaNotify,
		PendingCLAId:   do.PendingCLAId,
		ClaNotifyCount: do.ClaNotifyCount,
		ClaNotifyTime:  do.ClaNotifyTime,
	}
}

func (do *corpSigningDO) allManagers() []domain.Manager {
	v := do.toManagers()
	if do.Admin.isEmpty() {
		return v
	}

	return append(v, do.Admin.toManager())
}

func (do *corpSigningDO) toCorpSigning() domain.CorpSigning {
	return domain.CorpSigning{
		Id:   do.index(),
		Rep:  do.Rep.toRep(),
		Corp: do.Corp.toCorp(),
		Date: do.Date,
		Link: domain.LinkInfo{
			Id: do.LinkId,
			CLAInfo: domain.CLAInfo{
				CLAId:    do.CLAId,
				Language: dp.CreateLanguage(do.Language),
			},
		},
		Admin:          do.Admin.toManager(),
		HasPDF:         do.HasPDF,
		AllInfo:        do.AllInfo,
		Managers:       do.toManagers(),
		Employees:      do.toEmployeeSignings(),
		Version:        do.Version,
		PendingCLAId:   do.PendingCLAId,
		ClaNotifyCount: do.ClaNotifyCount,
		ClaNotifyTime:  do.ClaNotifyTime,
		Logs:           do.toCorpSigningLogs(),
	}
}

func (do *corpSigningDO) toEmployeeSignings() []domain.EmployeeSigning {
	es := make([]domain.EmployeeSigning, len(do.Employees))
	for i := range do.Employees {
		es[i] = do.Employees[i].toEmployeeSigning()
	}

	return es
}

func toEmployeeSigningDOs(employees []domain.EmployeeSigning) []employeeSigningDO {
	if employees == nil {
		return nil
	}

	result := make([]employeeSigningDO, len(employees))
	for i := range employees {
		result[i] = toEmployeeSigningDO(&employees[i])
	}
	return result
}

func (do *corpSigningDO) toManagers() []domain.Manager {
	ms := make([]domain.Manager, len(do.Managers))
	for i := range do.Managers {
		ms[i] = do.Managers[i].toManager()
	}

	return ms
}

func toManagerDOs(managers []domain.Manager) []managerDO {
	if managers == nil {
		return nil
	}

	result := make([]managerDO, len(managers))
	for i := range managers {
		result[i] = toManagerDO(&managers[i])
	}
	return result
}

// representative DO
type RepDO struct {
	Name  string `bson:"name"  json:"name"  required:"true"`
	Email string `bson:"email" json:"email" required:"true"`
}

func (do *RepDO) toRep() domain.Representative {
	return domain.Representative{
		Name:      dp.CreateName(do.Name),
		EmailAddr: dp.CreateEmailAddr(do.Email),
	}
}

func toRepDO(v *domain.Representative) RepDO {
	repDO := RepDO{}

	if v != nil {
		if v.Name != nil {
			repDO.Name = v.Name.Name()
		}
		if v.EmailAddr != nil {
			repDO.Email = v.EmailAddr.EmailAddr()
		}
	}

	return repDO
}

// corporation DO
type corpDO struct {
	Name    string   `bson:"name"     json:"name"      required:"true"`
	Domain  string   `bson:"domain"   json:"domain"    required:"true"`
	Domains []string `bson:"domains"  json:"domains"   required:"true"`
}

func (do *corpDO) toCorp() domain.Corporation {
	return domain.Corporation{
		Name:               dp.CreateCorpName(do.Name),
		AllEmailDomains:    do.Domains,
		PrimaryEmailDomain: do.Domain,
	}
}

func toCorpDO(v *domain.Corporation) corpDO {
	if v == nil {
		return corpDO{}
	}

	name := ""
	if v.Name != nil {
		name = v.Name.CorpName()
	}

	domain := ""
	if v.PrimaryEmailDomain != "" {
		domain = v.PrimaryEmailDomain
	}

	domains := []string{}
	if v.AllEmailDomains != nil {
		domains = v.AllEmailDomains
	}

	return corpDO{
		Name:    name,
		Domain:  domain,
		Domains: domains,
	}
}

type CorpSigningLogsDO struct {
	Logs []corpSigningLogDO `bson:"logs" json:"logs"`
}

type corpSigningLogDO struct {
	Date   string `bson:"date"   json:"date"`
	CLAId  string `bson:"cla_id" json:"cla_id"`
	Action string `bson:"action" json:"action"`
}

func toCorpSigningLogsDO(logs []domain.CorpSigningLog) CorpSigningLogsDO {
	var dos []corpSigningLogDO
	for _, v := range logs {
		dos = append(dos, toCorpSigningLogDO(v))
	}
	return CorpSigningLogsDO{dos}
}

func toCorpSigningLogDO(log domain.CorpSigningLog) corpSigningLogDO {
	return corpSigningLogDO{
		Date:   log.Date,
		CLAId:  log.CLAId,
		Action: log.Action,
	}
}

func (do *corpSigningLogDO) toCorpSigningLog() domain.CorpSigningLog {
	return domain.CorpSigningLog{
		Date:   do.Date,
		CLAId:  do.CLAId,
		Action: do.Action,
	}
}

func (do *corpSigningDO) toCorpSigningLogs() []domain.CorpSigningLog {
	var logs []domain.CorpSigningLog
	for _, v := range do.Logs {
		logs = append(logs, v.toCorpSigningLog())
	}
	return logs
}
