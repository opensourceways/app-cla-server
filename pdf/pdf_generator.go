package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type pdfGenerator struct {
	pdfOutDir string
	pythonBin string
	corp      []*corpSigningPDF
}

func (pg *pdfGenerator) generator(claLang string) *corpSigningPDF {
	for _, item := range pg.corp {
		if item.language == strings.ToLower(claLang) {
			return item
		}
	}
	return nil
}

func (pg *pdfGenerator) GenPDFForCorporationSigning(linkID, claFile string, signing *models.CorporationSigning, claFields []models.CLAField) (string, error) {
	corp := pg.generator(signing.CLALanguage)
	if corp == nil {
		return "", fmt.Errorf("unknown cla language:%s", signing.CLALanguage)
	}

	tempPdf := util.GenFilePath(pg.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, "_sig"))
	err := genSignaturePDF(corp, signing, claFields, tempPdf)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPdf)

	outfile := util.GenFilePath(pg.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, ""))
	if err := appendCorpPDFSignaturePage(pg.pythonBin, claFile, tempPdf, outfile); err != nil {
		return "", err
	}

	return outfile, nil
}

func genSignaturePDF(c *corpSigningPDF, signing *models.CorporationSigning, claFields []models.CLAField, outFile string) error {
	pdf := c.newPDF()

	pdf.AddPage()
	orders, titles := BuildCorpContact(claFields)
	c.addSignature(pdf, signing.Info, orders, titles)

	if !util.IsFileNotExist(outFile) {
		if err := os.Remove(outFile); err != nil {
			return err
		}
	}

	if err := c.end(pdf, outFile); err != nil {
		return fmt.Errorf("generate signing pdf of corp failed: %s", err.Error())
	}
	return nil
}

func appendCorpPDFSignaturePage(pythonBin, pdfFile, sigFile, outfile string) error {
	if util.IsFileNotExist(sigFile) {
		return fmt.Errorf("org signature file(%s) is not exist", sigFile)
	}

	// merge file
	cmd := exec.Command(pythonBin, "./util/merge_signature.py", "append", pdfFile, sigFile, outfile)
	if out, err := cmd.Output(); err != nil {
		return fmt.Errorf("append signature page of pdf failed: %s, %s", out, err.Error())
	}

	return nil
}

func BuildCorpContact(fields []models.CLAField) ([]string, map[string]string) {
	ids := make(sort.IntSlice, 0, len(fields))
	m := map[int]string{}
	mk := map[string]string{}

	for i := range fields {
		item := &fields[i]
		v, err := strconv.Atoi(item.ID)
		if err != nil {
			continue
		}

		ids = append(ids, v)
		m[v] = item.ID
		mk[item.ID] = item.Title
	}

	ids.Sort()

	r := make([]string, 0, len(ids))
	for _, k := range ids {
		r = append(r, m[k])
	}
	return r, mk
}

func genPDFFileName(linkID, email, other string) string {
	s := strings.ReplaceAll(util.EmailSuffix(email), ".", "_")
	return fmt.Sprintf("%s_%s%s.pdf", linkID, s, other)
}
