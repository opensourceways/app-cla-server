package pdf

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type pdfGenerator struct {
	pdfOutDir    string
	pdfOrgSigDir string
	pythonBin    string
	corp         *corpSigningPDF
}

func (this *pdfGenerator) GenPDFForCorporationSigning(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error) {
	tempPdf, err := genCorporPDFMissingSig(this.corp, orgCLA, signing, cla, this.pdfOutDir)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempPdf)

	lock := util.NewFileLock(
		util.LockedFilePath(this.pdfOrgSigDir, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID),
	)
	if err := lock.Lock(); err != nil {
		return "", fmt.Errorf("lock failed: %s", err.Error())
	}
	defer lock.Unlock()

	orgSigPdfFile := util.OrgSignaturePDFFILE(this.pdfOrgSigDir, orgCLA.ID)
	file := util.CorporCLAPDFFile(this.pdfOutDir, orgCLA.ID, signing.AdminEmail, "")
	if err := mergeCorporPDFSignaturePage(orgCLA.ID, this.pythonBin, tempPdf, orgSigPdfFile, file); err != nil {
		return "", err
	}

	return file, nil
}

func genCorporPDFMissingSig(c *corpSigningPDF, orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA, outDir string) (string, error) {
	project := orgCLA.OrgID
	if orgCLA.RepoID != "" {
		project = fmt.Sprintf("%s-%s", project, orgCLA.RepoID)
	}

	pdf := c.begin()

	// first page
	c.firstPage(pdf, fmt.Sprintf("The %s Project", project))
	c.welcome(pdf, project, orgCLA.OrgEmail)

	orders, keys, err := buildCorpContact(cla)
	if err != nil {
		return "", fmt.Errorf("build contact info of corp signing failed: %s", err.Error())
	}
	c.contact(pdf, signing.Info, orders, keys)

	c.declare(pdf, project)
	c.cla(pdf, cla.Text)

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

func buildCorpContact(cla *models.CLA) ([]string, map[string]string, error) {
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
