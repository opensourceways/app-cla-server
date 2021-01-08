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

func (this *pdfGenerator) GetBlankSignaturePath(claLang string) string {
	return util.GenFilePath(this.pdfOrgSigDir, strings.ToLower(claLang)+"_blank_signature.pdf")
}

func (this *pdfGenerator) GenPDFForCorporationSigning(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error) {
	tempPdf, err := genCorporPDFMissingSig(this.corp, orgCLA, signing, cla, this.pdfOutDir)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPdf)

	unlock, err := util.Lock(
		util.GenFilePath(
			this.pdfOrgSigDir,
			util.GenFileName(orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID),
		),
	)
	if err != nil {
		return "", fmt.Errorf("lock failed: %s", err.Error())
	}
	defer unlock()

	orgSigPdfFile := util.OrgSignaturePDFFILE(this.pdfOrgSigDir, orgCLA.ID)
	file := util.CorporCLAPDFFile(this.pdfOutDir, orgCLA.ID, signing.AdminEmail, "")
	if err := mergeCorporPDFSignaturePage(orgCLA.ID, this.pythonBin, tempPdf, orgSigPdfFile, file); err != nil {
		return "", err
	}

	return file, nil
}

func genCorporPDFMissingSig(c *corpSigningPDF, orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA, outDir string) (string, error) {
	pdf := c.begin()

	// first page
	c.firstPage(pdf, fmt.Sprintf("The Project of %s", orgCLA.OrgAlias))
	c.welcome(pdf, orgCLA.OrgAlias, orgCLA.OrgEmail)

	orders, titles := BuildCorpContact(cla)
	c.contact(pdf, signing.Info, orders, titles)

	c.declare(pdf)
	c.cla(pdf, cla.Text)
	c.projectURL(pdf, fmt.Sprintf("[1]. %s", util.ProjectURL(orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)))

	// second page
	c.secondPage(pdf, signing.Date)

	path := util.CorporCLAPDFFile(outDir, orgCLA.ID, signing.AdminEmail, "_missing_sig")
	if !util.IsFileNotExist(path) {
		os.Remove(path)
	}
	if err := c.end(pdf, path); err != nil {
		return "", fmt.Errorf("generate signing pdf of corp failed: %s", err.Error())
	}
	return path, nil
}

func mergeCorporPDFSignaturePage(orgCLAID, pythonBin, pdfFile, sigFile, outfile string) error {
	md5sum := ""
	var err error
	if !util.IsFileNotExist(sigFile) {
		if md5sum, err = util.Md5sumOfFile(sigFile); err != nil {
			return fmt.Errorf("calculate md5sum failed: %s", err.Error())
		}
	}

	// fetch signature, it will be returned when md5sum is not matched.
	signature, err := models.DownloadOrgSignatureByMd5(orgCLAID, md5sum)
	if err != nil {
		return fmt.Errorf("download org's signature failed: %s", err.Error())
	}

	if signature == nil {
		if md5sum == "" {
			return fmt.Errorf("the org's signature has not been uploaded")
		}
	} else {
		// write signature
		if err := ioutil.WriteFile(sigFile, signature, 0644); err != nil {
			return fmt.Errorf("write org's signature failed: %s", err.Error())
		}
	}

	// merge file
	cmd := exec.Command(
		pythonBin, "./util/merge-signature.py",
		pdfFile, sigFile, outfile,
	)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("merge signature page of pdf failed: %s", err.Error())
	}

	return nil
}

func BuildCorpContact(cla *models.CLA) ([]string, map[string]string) {
	ids := make(sort.IntSlice, 0, len(cla.Fields))
	m := map[int]string{}
	mk := map[string]string{}

	for _, item := range cla.Fields {
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
