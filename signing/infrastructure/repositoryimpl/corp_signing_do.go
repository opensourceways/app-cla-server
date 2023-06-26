package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldRep       = "rep"
	fieldCorp      = "corp"
	fieldName      = "name"
	fieldAdmin     = "admin"
	fieldEmail     = "email"
	fieldDomain    = "domain"
	fieldLinkId    = "link_id"
	fieldVersion   = "version"
	fieldManagers  = "managers"
	fieldEmployees = "employees"
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

	Admin     managerDO           `bson:"admin"         json:"admin"`
	Managers  []managerDO         `bson:"managers"      json:"managers"`
	Employees []employeeSigningDO `bson:"employees"     json:"employees"`
	Version   int                 `bson:"version"       json:"-"`
}

func (do *corpSigningDO) toDoc() (bson.M, error) {
	return genDoc(do)
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
		Id:        do.Id.Hex(),
		Date:      do.Date,
		Rep:       rep,
		Corp:      corp,
		AllInfo:   do.AllInfo,
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
