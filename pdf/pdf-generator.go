package pdf

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type pdfGenerator struct {
	pdfOutDir    string
	pdfOrgSigDir string
	pythonBin    string
	corp         *corpSigningPDF
}

func (this *pdfGenerator) LangSupported() map[string]bool {
	return map[string]bool{this.corp.language: true}
}

func (this *pdfGenerator) GetBlankSignaturePath(claLang string) string {
	return util.GenFilePath(this.pdfOrgSigDir, strings.ToLower(claLang)+"_blank_signature.pdf")
}

func (this *pdfGenerator) GenPDFForCorporationSigning(linkID, orgSignatureFile, claFile string, orgInfo *models.OrgInfo, signing *models.CorporationSigning, claFields []models.CLAField) (string, error) {
	tempPdf := util.GenFilePath(this.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, "_missing_sig"))
	err := genCorporPDFMissingSig(this.corp, orgInfo, signing, claFields, claFile, tempPdf)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPdf)

	outfile := util.GenFilePath(this.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, ""))
	if err := mergeCorporPDFSignaturePage(this.pythonBin, tempPdf, orgSignatureFile, outfile); err != nil {
		return "", err
	}

	return outfile, nil
}

func genCorporPDFMissingSig(c *corpSigningPDF, orgInfo *models.OrgInfo, signing *models.CorporationSigning, claFields []models.CLAField, claFile, outFile string) error {
	text, err := ioutil.ReadFile(claFile)
	if err != nil {
		return fmt.Errorf("failed to read cla file(%s): %s", claFile, err.Error())
	}

	pdf := c.begin()

	// first page
	c.firstPage(pdf, orgInfo.OrgAlias)
	c.welcome(pdf, orgInfo.OrgAlias, orgInfo.OrgEmail)

	orders, titles := BuildCorpContact(claFields)
	c.contact(pdf, signing.Info, orders, titles)

	c.declare(pdf)
	c.cla(pdf, string(text))
	c.projectURL(pdf, fmt.Sprintf("[1]. %s", orgInfo.ProjectURL()))

	// second page
	c.secondPage(pdf, signing.Date)

	if !util.IsFileNotExist(outFile) {
		os.Remove(outFile)
	}
	if err := c.end(pdf, outFile); err != nil {
		return fmt.Errorf("generate signing pdf of corp failed: %s", err.Error())
	}
	return nil
}

func mergeCorporPDFSignaturePage(pythonBin, pdfFile, sigFile, outfile string) error {
	if util.IsFileNotExist(sigFile) {
		return fmt.Errorf("org signature file(%s) is not exist", sigFile)
	}

	// merge file
	cmd := exec.Command(pythonBin, "./util/merge-signature.py", pdfFile, sigFile, outfile)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("merge signature page of pdf failed: %s", err.Error())
	}

	return nil
}

func BuildCorpContact(fields []models.CLAField) ([]string, map[string]string) {
	ids := make(sort.IntSlice, 0, len(fields))
	m := map[int]string{}
	mk := map[string]string{}

	for _, item := range fields {
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
