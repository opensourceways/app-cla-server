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
	corp         []*corpSigningPDF
}

func (this *pdfGenerator) LangSupported() map[string]bool {
	v := map[string]bool{}
	for _, item := range this.corp {
		v[item.language] = true
	}
	return v
}

func (this *pdfGenerator) GetBlankSignaturePath(claLang string) string {
	return util.GenFilePath(this.pdfOrgSigDir, strings.ToLower(claLang)+"_blank_signature.pdf")
}

func (this *pdfGenerator) generator(claLang string) *corpSigningPDF {
	for _, item := range this.corp {
		if item.language == strings.ToLower(claLang) {
			return item
		}
	}
	return nil
}

func (this *pdfGenerator) GenPDFForCorporationSigning1(linkID, claFile string, signing *models.CorporationSigning, claFields []models.CLAField) (string, error) {
	corp := this.generator(signing.CLALanguage)
	if corp == nil {
		return "", fmt.Errorf("unknown cla language:%s", signing.CLALanguage)
	}

	outFile := util.GenFilePath(this.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, ""))
	err := genCorporPDF(corp, signing, claFields, claFile, outFile)
	return outFile, err
}

func (this *pdfGenerator) GenPDFForCorporationSigning(linkID, claFile string, signing *models.CorporationSigning, claFields []models.CLAField) (string, error) {
	corp := this.generator(signing.CLALanguage)
	if corp == nil {
		return "", fmt.Errorf("unknown cla language:%s", signing.CLALanguage)
	}

	tempPdf := util.GenFilePath(this.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, "_sig"))
	err := genSignaturePDF(corp, signing, claFields, tempPdf)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPdf)

	outfile := util.GenFilePath(this.pdfOutDir, genPDFFileName(linkID, signing.AdminEmail, ""))
	if err := appendCorpPDFSignaturePage(this.pythonBin, claFile, tempPdf, outfile); err != nil {
		return "", err
	}

	return outfile, nil
}

func genSignaturePDF(c *corpSigningPDF, signing *models.CorporationSigning, claFields []models.CLAField, outFile string) error {
	pdf := c.begin()

	pdf.AddPage()
	orders, titles := BuildCorpContact(claFields)
	c.addSignature(pdf, signing.Info, orders, titles)

	if !util.IsFileNotExist(outFile) {
		os.Remove(outFile)
	}
	if err := c.end(pdf, outFile); err != nil {
		return fmt.Errorf("generate signing pdf of corp failed: %s", err.Error())
	}
	return nil
}

func genCorporPDF(c *corpSigningPDF, signing *models.CorporationSigning, claFields []models.CLAField, claFile, outFile string) error {
	text, err := ioutil.ReadFile(claFile)
	if err != nil {
		return fmt.Errorf("failed to read cla file(%s): %s", claFile, err.Error())
	}

	pdf := c.begin()

	// first page
	pdf.AddPage()

	c.cla(pdf, string(text))

	// second page
	pdf.AddPage()
	orders, titles := BuildCorpContact(claFields)
	c.addSignature(pdf, signing.Info, orders, titles)

	if !util.IsFileNotExist(outFile) {
		os.Remove(outFile)
	}
	if err := c.end(pdf, outFile); err != nil {
		return fmt.Errorf("generate signing pdf of corp failed: %s", err.Error())
	}
	return nil
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

func appendCorpPDFSignaturePage(pythonBin, pdfFile, sigFile, outfile string) error {
	if util.IsFileNotExist(sigFile) {
		return fmt.Errorf("org signature file(%s) is not exist", sigFile)
	}

	// merge file
	cmd := exec.Command(pythonBin, "./util/merge-signature.py", "append", pdfFile, sigFile, outfile)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("append signature page of pdf failed: %s", err.Error())
	}

	return nil
}
func mergeCorpPDFSignaturePage(pythonBin, pdfFile, sigFile, outfile string) error {
	if util.IsFileNotExist(sigFile) {
		return fmt.Errorf("org signature file(%s) is not exist", sigFile)
	}

	// merge file
	cmd := exec.Command(pythonBin, "./util/merge-signature.py", "merge", pdfFile, sigFile, outfile)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("merge signature page of pdf failed: %s", err.Error())
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
