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

func (do *corpSigningDO) toCorpSigningSummary(cs *repository.CorpSigningSummary) (err error) {
	rep, err := do.Rep.toRep()
	if err != nil {
		return
	}

	corp, err := do.Corp.toCorp()
	if err != nil {
		return
	}

	admin, err := do.Admin.toManager()
	if err != nil {
		return
	}

	*cs = repository.CorpSigningSummary{
		Id:     do.index(),
		Date:   do.Date,
		HasPDF: do.HasPDF,
		Rep:    rep,
		Corp:   corp,
		Admin:  admin,
	}

	cs.Link.Id = do.LinkId
	cs.Link.Language, err = dp.NewLanguage(do.Language)

	return
}

func (do *corpSigningDO) allManagers() ([]domain.Manager, error) {
	v, err := do.toManagers()
	if err != nil || do.Admin.isEmpty() {
		return v, err
	}

	admin, err := do.Admin.toManager()
	if err != nil {
		return nil, err
	}

	return append(v, admin), nil
}

func (do *corpSigningDO) toCorpSigning(cs *domain.CorpSigning) (err error) {
	rep, err := do.Rep.toRep()
	if err != nil {
		return
	}

	corp, err := do.Corp.toCorp()
	if err != nil {
		return
	}

	es, err := do.toEmployeeSignings()
	if err != nil {
		return
	}

	admin, err := do.Admin.toManager()
	if err != nil {
		return
	}

	managers, err := do.toManagers()
	if err != nil {
		return
	}

	*cs = domain.CorpSigning{
		Id:        do.index(),
		Date:      do.Date,
		Rep:       rep,
		Corp:      corp,
		AllInfo:   do.AllInfo,
		HasPDF:    do.HasPDF,
		Admin:     admin,
		Managers:  managers,
		Employees: es,
		Version:   do.Version,
	}

	cs.Link.Id = do.LinkId
	cs.Link.CLAId = do.CLAId
	cs.Link.Language, err = dp.NewLanguage(do.Language)

	return
}

func (do *corpSigningDO) toEmployeeSignings() (es []domain.EmployeeSigning, err error) {
	es = make([]domain.EmployeeSigning, len(do.Employees))

	for i := range do.Employees {
		if err = do.Employees[i].toEmployeeSigning(&es[i]); err != nil {
			return
		}
	}

	return
}

func (do *corpSigningDO) toManagers() (ms []domain.Manager, err error) {
	ms = make([]domain.Manager, len(do.Managers))

	for i := range do.Managers {
		if ms[i], err = do.Managers[i].toManager(); err != nil {
			return
		}
	}

	return
}

// representative DO
type RepDO struct {
	Name  string `bson:"name"  json:"name"  required:"true"`
	Email string `bson:"email" json:"email" required:"true"`
}

func (do *RepDO) toRep() (rep domain.Representative, err error) {
	if rep.Name, err = dp.NewName(do.Name); err != nil {
		return
	}

	rep.EmailAddr, err = dp.NewEmailAddr(do.Email)

	return
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

func (do *corpDO) toCorp() (c domain.Corporation, err error) {
	if c.Name, err = dp.NewCorpName(do.Name); err != nil {
		return
	}

	c.PrimaryEmailDomain = do.Domain
	c.AllEmailDomains = do.Domains

	return
}

func toCorpDO(v *domain.Corporation) corpDO {
	return corpDO{
		Name:    v.Name.CorpName(),
		Domain:  v.PrimaryEmailDomain,
		Domains: v.AllEmailDomains,
	}
}
