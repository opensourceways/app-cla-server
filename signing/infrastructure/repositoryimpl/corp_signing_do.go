package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

const (
	fieldPDF       = "pdf"
	fieldRep       = "rep"
	fieldDate      = "date"
	fieldCorp      = "corp"
	fieldName      = "name"
	fieldLang      = "lang"
	fieldAdmin     = "admin"
	fieldEmail     = "email"
	fieldHasPDF    = "has_pdf"
	fieldLinkId    = "link_id"
	fieldDomain    = "domain"
	fieldDomains   = "domains"
	fieldDeleted   = "deleted"
	fieldVersion   = "version"
	fieldManagers  = "managers"
	fieldEmployees = "employees"
	fieldTriggered = "triggered"
)

func toCorpSigningDO(v *domain.CorpSigning) corpSigningDO {
	link := &v.Link

	return corpSigningDO{
		Date:     v.Date,
		CLAId:    link.CLAId,
		LinkId:   link.Id,
		Language: link.Language.Language(),
		Rep:      toRepDO(&v.Rep),
		Corp:     toCorpDO(&v.Corp),
		AllInfo:  v.AllInfo,
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

	// uploading pdf or adding email domain will trigger individual signing checking
	// which will delete the one that belongs to a corp.
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
		Admin:  do.Admin.toManager(),
		HasPDF: do.HasPDF,
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
		Admin:     do.Admin.toManager(),
		HasPDF:    do.HasPDF,
		AllInfo:   do.AllInfo,
		Managers:  do.toManagers(),
		Employees: do.toEmployeeSignings(),
		Version:   do.Version,
	}
}

func (do *corpSigningDO) toEmployeeSignings() []domain.EmployeeSigning {
	es := make([]domain.EmployeeSigning, len(do.Employees))

	for i := range do.Employees {
		do.Employees[i].toEmployeeSigning(&es[i])
	}

	return es
}

func (do *corpSigningDO) toManagers() []domain.Manager {
	ms := make([]domain.Manager, len(do.Managers))
	for i := range do.Managers {
		ms[i] = do.Managers[i].toManager()
	}

	return ms
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
	return RepDO{
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

func (do *corpDO) toCorp() domain.Corporation {
	return domain.Corporation{
		Name:               dp.CreateCorpName(do.Name),
		AllEmailDomains:    do.Domains,
		PrimaryEmailDomain: do.Domain,
	}
}

func toCorpDO(v *domain.Corporation) corpDO {
	return corpDO{
		Name:    v.Name.CorpName(),
		Domain:  v.PrimaryEmailDomain,
		Domains: v.AllEmailDomains,
	}
}
