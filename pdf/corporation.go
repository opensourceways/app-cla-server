package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

func (this *pdfGenerator) GenCLAPDFForCorporation(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error) {
	orgSigPdfFile := util.OrgSignaturePDFFILE(this.pdfOrgSigDir, orgCLA.ID)
	if util.IsFileNotExist(orgSigPdfFile) {
		return "", fmt.Errorf("Failed to generate pdf for corporation signing: the org signature pdf file is not exist")
	}

	tempPdf, err := this.genCorporPDFMissingSig(orgCLA, signing, cla)
	if err != nil {
		return "", err
	}

	file := util.CorporCLAPDFFile(this.pdfOutDir, orgCLA.ID, signing.AdminEmail, "")
	if err := this.mergeCorporPDFSignaturePage(tempPdf, orgSigPdfFile, file); err != nil {
		return "", err
	}

	os.Remove(tempPdf)

	return file, nil
}

func (this *pdfGenerator) genCorporPDFMissingSig(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error) {
	c := this.corporation

	project := orgCLA.OrgID
	if orgCLA.RepoID != "" {
		project = fmt.Sprintf("%s-%s", project, orgCLA.RepoID)
	}

	pdf := c.begin()

	// first page
	c.firstPage(pdf, fmt.Sprintf("The %s Project", project))
	c.welcome(pdf, project, orgCLA.OrgEmail)

	orders, keys, err := buildCorporContact(cla)
	if err != nil {
		return "", err
	}
	c.contact(pdf, signing.Info, orders, keys)

	c.declare(pdf, project)
	c.cla(pdf, cla.Text)

	// second page
	c.secondPage(pdf, signing.Date)

	path := util.CorporCLAPDFFile(this.pdfOutDir, orgCLA.ID, signing.AdminEmail, "_missing_sig")
	if err := c.end(pdf, path); err != nil {
		return "", err
	}
	return path, nil
}

func (this *pdfGenerator) mergeCorporPDFSignaturePage(pdfFile, sigFile, outfile string) error {
	cmd := exec.Command(
		this.pythonBin, "./util/merge-signature.py",
		pdfFile, sigFile, outfile,
	)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to merge signature page for corporation pdf: %s", err.Error())
	}

	return nil
}

func buildCorporContact(cla *models.CLA) ([]string, map[string]string, error) {
	ids := make(sort.IntSlice, 0, len(cla.Fields))
	m := map[int]string{}
	mk := map[string]string{}

	for _, item := range cla.Fields {
		v, err := strconv.Atoi(item.ID)
		if err != nil {
			return nil, nil, err
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
	return r, mk, nil
}
