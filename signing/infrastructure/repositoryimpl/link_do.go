package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldOrg      = "org"
	fieldCLAs     = "clas"
	fieldCLANum   = "cla_num"
	fieldRemoved  = "removed"
	fieldOrgAlias = "org_alias"
	fieldPlatform = "platform"
)

func toLinkDO(v *domain.Link) linkDO {
	do := linkDO{
		Id:        v.Id,
		Org:       toOrgInfoDO(&v.Org),
		Email:     toEmailInfoDO(&v.Email),
		Submitter: v.Submitter,
		CLANum:    v.CLANum,
	}

	clas := make([]claDO, len(v.CLAs))
	for i := range v.CLAs {
		clas[i] = toCLADO(&v.CLAs[i])
	}

	do.CLAs = clas

	return do
}

type linkDO struct {
	Id          string      `bson:"id"         json:"id"          required:"true"`
	Org         orgInfoDO   `bson:"org"        json:"org"         required:"true"`
	Email       emailInfoDO `bson:"email"      json:"email"       required:"true"`
	Submitter   string      `bson:"submitter"  json:"submitter"   required:"true"`
	CLAs        []claDO     `bson:"clas"       json:"clas"`
	CLANum      int         `bson:"cla_num"    json:"cla_num"`
	Version     int         `bson:"version"    json:"-"`
	Deleted     bool        `bson:"deleted"    json:"deleted"`
	RemovedCLAs []claDO     `bson:"removed"    json:"removed"`
}

func (do *linkDO) toLink(link *domain.Link) (err error) {
	e, err := do.Email.toEmailInfo()
	if err != nil {
		return
	}

	clas := make([]domain.CLA, len(do.CLAs))
	for i := range do.CLAs {
		if clas[i], err = do.CLAs[i].toCLA(); err != nil {
			return
		}
	}

	*link = domain.Link{
		Id:        do.Id,
		Org:       do.Org.toOrgInfo(),
		Email:     e,
		CLAs:      clas,
		Submitter: do.Submitter,
		CLANum:    do.CLANum,
		Version:   do.Version,
	}

	return
}

func (do *linkDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

// orgInfoDO
type orgInfoDO struct {
	Alias      string `bson:"org_alias" json:"org_alias"`
	ProjectURL string `bson:"project"   json:"project"    required:"true"`
}

func (do *orgInfoDO) toOrgInfo() domain.OrgInfo {
	return domain.OrgInfo{
		Alias:      do.Alias,
		ProjectURL: do.ProjectURL,
	}
}

func toOrgInfoDO(v *domain.OrgInfo) orgInfoDO {
	return orgInfoDO{
		Alias:      v.Alias,
		ProjectURL: v.ProjectURL,
	}
}

// emailInfoDO
type emailInfoDO struct {
	Addr     string `bson:"addr"     json:"addr"      required:"true"`
	Platform string `bson:"platform" json:"platform"  required:"true"`
}

func (do *emailInfoDO) toEmailInfo() (v domain.EmailInfo, err error) {
	if v.Addr, err = dp.NewEmailAddr(do.Addr); err != nil {
		return
	}

	v.Platform = do.Platform

	return
}

func toEmailInfoDO(v *domain.EmailInfo) emailInfoDO {
	return emailInfoDO{
		Addr:     v.Addr.EmailAddr(),
		Platform: v.Platform,
	}
}

// fieldDO
type fieldDO struct {
	Id       string `bson:"id"       json:"id"     required:"true"`
	Type     string `bson:"type"     json:"type"   required:"true"`
	Desc     string `bson:"desc"     json:"desc,omitempty"`
	Title    string `bson:"title"    json:"title"  required:"true"`
	Required bool   `bson:"required" json:"required"`
}

func (do *fieldDO) toField(t dp.CLAType) (v domain.Field, err error) {
	v.Id = do.Id
	v.Type = do.Type
	v.Desc = do.Desc
	v.Title = do.Title
	v.Required = do.Required

	return
}

func toFieldDO(v *domain.Field) fieldDO {
	return fieldDO{
		Id:       v.Id,
		Type:     v.Type,
		Desc:     v.Desc,
		Title:    v.Title,
		Required: v.Required,
	}
}

// claDO
type claDO struct {
	Id       string    `bson:"id"      json:"id"     required:"true"`
	URL      string    `bson:"url"     json:"url"    required:"true"`
	Type     string    `bson:"type"    json:"type"   required:"true"`
	Fields   []fieldDO `bson:"fields"  json:"fields,omitempty"`
	Language string    `bson:"lang"    json:"lang"   required:"true"`
}

func (do *claDO) toCLA() (cla domain.CLA, err error) {
	if cla.URL, err = dp.NewURL(do.URL); err != nil {
		return
	}

	if cla.Type, err = dp.NewCLAType(do.Type); err != nil {
		return
	}

	if cla.Language, err = dp.NewLanguage(do.Language); err != nil {
		return
	}

	cla.Id = do.Id

	fields := make([]domain.Field, len(do.Fields))

	for i := range do.Fields {
		if fields[i], err = do.Fields[i].toField(cla.Type); err != nil {
			return
		}
	}

	cla.Fields = fields

	return
}

func (do *claDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func toCLADO(v *domain.CLA) claDO {
	fields := make([]fieldDO, len(v.Fields))
	for i := range v.Fields {
		fields[i] = toFieldDO(&v.Fields[i])
	}

	return claDO{
		Id:       v.Id,
		URL:      v.URL.URL(),
		Type:     v.Type.CLAType(),
		Fields:   fields,
		Language: v.Language.Language(),
	}
}
