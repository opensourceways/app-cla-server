package adapter

import (
	"errors"
	"strconv"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

func NewCLAAdapter(
	s app.CLAService,
	maxSizeOfCLAContent int,
	fileTypeOfCLAContent string,
) *claAdatper {
	return &claAdatper{
		s:                    s,
		maxSizeOfCLAContent:  maxSizeOfCLAContent,
		fileTypeOfCLAContent: fileTypeOfCLAContent,
	}
}

type claAdatper struct {
	s                    app.CLAService
	maxSizeOfCLAContent  int
	fileTypeOfCLAContent string
}

// Remove
func (adapter *claAdatper) Remove(linkId, claId string) models.IModelError {
	err := adapter.s.Remove(domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})

	if err != nil {
		return toModelError(err)
	}

	return nil
}

// List
func (adapter *claAdatper) List(linkId string) (dbmodels.CLAOfLink, models.IModelError) {
	individuals, corps, err := adapter.s.List(linkId)
	if err != nil {
		return dbmodels.CLAOfLink{}, toModelError(err)
	}

	return dbmodels.CLAOfLink{
		IndividualCLAs: adapter.toCLADetail(individuals),
		CorpCLAs:       adapter.toCLADetail(corps),
	}, nil
}

func (adapter *claAdatper) toCLADetail(v []app.CLADTO) []dbmodels.CLADetail {
	r := make([]dbmodels.CLADetail, len(v))

	for i := range v {
		item := &v[1]
		r[i].Language = item.Language
		r[i].URL = item.URL
	}

	return r
}

// CLALocalFilePath
func (adapter *claAdatper) CLALocalFilePath(linkId, claId string) string {
	return adapter.s.CLALocalFilePath(domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})
}

// Add
func (adapter *claAdatper) Add(linkId string, opt *models.CLACreateOpt, applyTo string) models.IModelError {
	cmd, err := adapter.cmdToAddCLA(opt, applyTo)
	if err != nil {
		return toModelError(err)
	}

	if err := adapter.s.Add(linkId, &cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *claAdatper) cmdToAddCLA(opt *models.CLACreateOpt, applyTo string) (
	cmd app.CmdToAddCLA, err error,
) {
	cmd.Text, err = util.DownloadFile(
		opt.URL, adapter.fileTypeOfCLAContent, adapter.maxSizeOfCLAContent,
	)
	if err != nil {
		return
	}

	if cmd.URL, err = dp.NewURL(opt.URL); err != nil {
		return
	}

	if cmd.Type, err = dp.NewCLAType(applyTo); err != nil {
		return
	}

	if cmd.Language, err = dp.NewLanguage(opt.Language); err != nil {
		return
	}

	cmd.Fields, err = adapter.toFields(cmd.Type, opt.Fields)

	return
}

func (adapter *claAdatper) toFields(claType dp.CLAType, fields []dbmodels.Field) (r []domain.Field, err error) {
	if len(fields) == 0 {
		err = errors.New("no fields")

		return
	}

	f := dp.NewCorpCLAFieldType
	if dp.IsCLATypeIndividual(claType) {
		f = dp.NewIndividualCLAFieldType
	}

	m := map[string]bool{}

	r = make([]domain.Field, len(fields))
	for i := range fields {
		item := &fields[i]

		m[item.Type] = true

		if r[i], err = adapter.toField(item, f); err != nil {
			return
		}
	}

	if len(m) != len(fields) {
		err = errors.New("duplicate fields")
	}

	return
}

func (adapter *claAdatper) toField(
	opt *dbmodels.Field,
	f func(string) (dp.CLAFieldType, error),
) (domain.Field, error) {
	t, err := f(opt.Type)
	if err != nil {
		return domain.Field{}, err
	}

	if _, err := strconv.Atoi(opt.ID); err != nil {
		return domain.Field{}, errors.New("invalid field id")
	}

	return domain.Field{
		Id:       opt.ID,
		Desc:     opt.Description,
		Type:     t,
		Title:    opt.Title,
		Required: opt.Required,
	}, nil
}
