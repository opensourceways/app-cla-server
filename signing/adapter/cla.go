package adapter

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
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
	pythonBin string,
	pythonRetryTimes int,
	pdfOutDir string,
) *claAdatper {
	return &claAdatper{
		s:                    s,
		claPDFSource:         claPDFSource,
		maxSizeOfCLAContent:  maxSizeOfCLAContent,
		fileTypeOfCLAContent: fileTypeOfCLAContent,
		pythonBin:            pythonBin,
		pythonRetryTimes:     pythonRetryTimes,
		pdfOutDir:            pdfOutDir,
	}
}

type claAdatper struct {
	s                    app.CLAService
	claPDFSource         []string
	maxSizeOfCLAContent  int
	fileTypeOfCLAContent string
	pythonBin            string
	pythonRetryTimes     int
	pdfOutDir            string
}

// Remove
func (adapter *claAdatper) Remove(userId, linkId, claId string) models.IModelError {
	err := adapter.s.Remove(&app.CmdToRemoveCLA{
		UserId: userId,
		CLAIndex: domain.CLAIndex{
			LinkId: linkId,
			CLAId:  claId,
		},
	})

	if err != nil {
		return toModelError(err)
	}

	return nil
}

// List
func (adapter *claAdatper) List(userId, linkId string) (models.CLAOfLink, models.IModelError) {
	individuals, corps, err := adapter.s.List(userId, linkId)
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
		r[i].UpdatedAt = formatCLADate(item.UpdatedAt)
	}

	return r
}

func formatCLADate(unix int64) string {
	if unix <= 0 {
		return time.Unix(0, 0).UTC().Format("2006-01-02")
	}
	return time.Unix(unix, 0).UTC().Format("2006-01-02")
}

// CLALocalFilePath
func (adapter *claAdatper) CLALocalFilePath(linkId, claId string) string {
	return adapter.s.CLALocalFilePath(domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})
}

// Add
func (adapter *claAdatper) Add(userId, linkId string, opt *models.CLACreateOpt) models.IModelError {
	cmd, err := adapter.cmdToAddCLA(userId, linkId, opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err := adapter.s.Add(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *claAdatper) Update(userId, linkId string, opt *models.CLAUpdateOpt) models.IModelError {
	cmd, err := adapter.cmdToUpdateCLA(userId, linkId, opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err = adapter.s.Update(&cmd); err != nil {
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

func (adapter *claAdatper) cmdToAddCLA(userId, linkId string, opt *models.CLACreateOpt) (
	cmd app.CmdToAddCLA, err error,
) {
	cmd.UserId = userId
	cmd.LinkId = linkId

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

func (adapter *claAdatper) cmdToUpdateCLA(userId, linkId string, opt *models.CLAUpdateOpt,
) (cmd app.CmdToUpdateCLA, err error) {
	cmd.UserId = userId
	cmd.LinkId = linkId

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

	cmd.Language, err = dp.NewLanguage(opt.Language)

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

// CorpCLATemplatePDF
func (adapter *claAdatper) CorpCLATemplatePDF(userId, linkId, claId string) (string, models.IModelError) {
	detail, err := adapter.s.CorpCLADetail(userId, linkId, claId)
	if err != nil {
		return "", toModelError(err)
	}

	gen := pdf.GetPDFGenerator()
	if gen == nil {
		return "", models.NewModelError(models.ErrGenTemplatePDFFailed, errors.New("pdf generator not initialized"))
	}

	outfile, err := gen.GenCLATemplatePDF(linkId, detail.LocalFile, detail.Language, toCLAFields(detail.Fileds))
	if err != nil {
		return "", models.NewModelError(models.ErrGenTemplatePDFFailed, err)
	}

	return outfile, nil
}

func toCLAFields(fields []domain.Field) []models.CLAField {
	r := make([]models.CLAField, len(fields))
	for i := range fields {
		item := &fields[i]
		r[i] = models.CLAField{
			ID:          item.Id,
			Title:       item.Title,
			Type:        item.Type,
			Description: item.Desc,
			Required:    item.Required,
		}
	}
	return r
}

// DiffPreview
func (adapter *claAdatper) DiffPreview(userId, linkId string, opt *models.CLADiffPreviewOpt) (
	models.CLADiffPreviewResult, models.IModelError,
) {
	if !adapter.isAllowedPDFSource(opt.URL) {
		return models.CLADiffPreviewResult{}, models.NewModelError(
			models.ErrNotAllowedCLAPDFSource, errors.New("not allowed cla pdf source"),
		)
	}

	claType, err := dp.NewCLAType(opt.Type)
	if err != nil {
		return models.CLADiffPreviewResult{}, errBadRequestParameter(err)
	}

	language, err := dp.NewLanguage(opt.Language)
	if err != nil {
		return models.CLADiffPreviewResult{}, errBadRequestParameter(err)
	}

	content, err := util.DownloadFile(opt.URL, adapter.fileTypeOfCLAContent, adapter.maxSizeOfCLAContent)
	if err != nil {
		return models.CLADiffPreviewResult{}, errBadRequestParameter(err)
	}

	newFile, err := util.WriteToTempFile("", "cla_diff_new_*.pdf", content)
	if err != nil {
		return models.CLADiffPreviewResult{}, models.NewModelError(models.ErrSystemError, err)
	}
	defer os.Remove(newFile)

	_, oldFile, err := adapter.s.DiffPreviewCLA(userId, linkId, claType, language)
	if err != nil {
		return models.CLADiffPreviewResult{}, toModelError(err)
	}

	if oldFile == "" {
		return models.CLADiffPreviewResult{IsNew: true, DiffHTML: ""}, nil
	}

	outHtml, err := ioutil.TempFile("", "cla_diff_out_*.html")
	if err != nil {
		return models.CLADiffPreviewResult{}, models.NewModelError(models.ErrGenDiffFailed, err)
	}
	outHtmlPath := outHtml.Name()
	outHtml.Close()
	defer os.Remove(outHtmlPath)

	if err := adapter.runGenerateDiff(oldFile, newFile, outHtmlPath); err != nil {
		return models.CLADiffPreviewResult{}, models.NewModelError(models.ErrGenDiffFailed, err)
	}

	html, err := ioutil.ReadFile(outHtmlPath)
	if err != nil {
		return models.CLADiffPreviewResult{}, models.NewModelError(models.ErrGenDiffFailed, err)
	}

	return models.CLADiffPreviewResult{IsNew: false, DiffHTML: string(html)}, nil
}

func (adapter *claAdatper) runGenerateDiff(oldFile, newFile, outFile string) error {
	times := adapter.pythonRetryTimes
	if times <= 0 {
		times = 3
	}

	var lastErr error
	for i := 0; i < times; i++ {
		cmd := exec.Command(adapter.pythonBin, "./util/generate_diff.py", oldFile, newFile, outFile)
		out, err := cmd.Output()
		if err == nil {
			return nil
		}
		lastErr = errors.New(string(out) + err.Error())
	}
	return lastErr
}
