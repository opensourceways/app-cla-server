package adapter

import (
	"errors"
	"strconv"
	"strings"

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
	claPDFSource []string,
) *claAdatper {
	return &claAdatper{
		s:                    s,
		claPDFSource:         claPDFSource,
		maxSizeOfCLAContent:  maxSizeOfCLAContent,
		fileTypeOfCLAContent: fileTypeOfCLAContent,
	}
}

type claAdatper struct {
	s                    app.CLAService
	claPDFSource         []string
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
func (adapter *claAdatper) List(linkId string) (models.CLAOfLink, models.IModelError) {
	individuals, corps, err := adapter.s.List(linkId)
	if err != nil {
		return models.CLAOfLink{}, toModelError(err)
	}

	return models.CLAOfLink{
		IndividualCLAs: adapter.toCLADetail(individuals),
		CorpCLAs:       adapter.toCLADetail(corps),
	}, nil
}

func (adapter *claAdatper) toCLADetail(v []app.CLADTO) []models.CLADetail {
	r := make([]models.CLADetail, len(v))

	for i := range v {
		item := &v[i]

		r[i].URL = item.URL
		r[i].CLAId = item.Id
		r[i].Language = item.Language
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
func (adapter *claAdatper) Add(linkId string, opt *models.CLACreateOpt) models.IModelError {
	cmd, err := adapter.cmdToAddCLA(opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err := adapter.s.Add(linkId, &cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *claAdatper) isAllowedPDFSource(url string) bool {
	for _, item := range adapter.claPDFSource {
		if strings.HasPrefix(url, item) {
			return true
		}
	}

	return false
}

func (adapter *claAdatper) cmdToAddCLA(opt *models.CLACreateOpt) (
	cmd app.CmdToAddCLA, err error,
) {
	if !adapter.isAllowedPDFSource(opt.URL) {
		err = errors.New("not allowed cla pdf source")

		return
	}

	cmd.Text, err = util.DownloadFile(
		opt.URL, adapter.fileTypeOfCLAContent, adapter.maxSizeOfCLAContent,
	)
	if err != nil {
		return
	}

	if cmd.URL, err = dp.NewURL(opt.URL); err != nil {
		return
	}

	if cmd.Type, err = dp.NewCLAType(opt.Type); err != nil {
		return
	}

	if cmd.Language, err = dp.NewLanguage(opt.Language); err != nil {
		return
	}

	cmd.Fields, err = adapter.toFields(cmd.Type, cmd.Language, opt.Fields)

	return
}

func (adapter *claAdatper) toFields(claType dp.CLAType, lang dp.Language, fields []models.CLAFieldCreateOpt) (
	r []domain.Field, err error,
) {
	if len(fields) == 0 {
		err = errors.New("no fields")

		return
	}

	all := dp.GetCLAFileds(claType, lang)
	allMap := make(map[string]*dp.CLAField, len(all))
	for i := range all {
		item := &all[i]
		allMap[item.Type] = item
	}

	m := map[string]bool{}

	r = make([]domain.Field, len(fields))
	for i := range fields {
		item := &fields[i]

		if m[item.Type] {
			err = errors.New("duplicate fields")

			return
		}
		m[item.Type] = true

		if r[i], err = adapter.toField(item, allMap); err != nil {
			return
		}
	}

	return
}

func (adapter *claAdatper) toField(opt *models.CLAFieldCreateOpt, all map[string]*dp.CLAField) (
	domain.Field, error,
) {
	field, ok := all[opt.Type]
	if !ok {
		return domain.Field{}, errors.New("invalid field")
	}

	if _, err := strconv.Atoi(opt.ID); err != nil {
		return domain.Field{}, errors.New("invalid field id")
	}

	return domain.Field{
		Id:       opt.ID,
		Required: opt.Required,
		CLAField: *field,
	}, nil
}
