package models

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CLACreateOption struct {
	dbmodels.CLAInfo

	content []byte
}

type CorpCLACreateOption struct {
	OrgSignature []byte `json:"org_signature"`
	*CLACreateOption
}

func (this *CorpCLACreateOption) AddCLA(orgRepo *dbmodels.OrgRepo) error {
	cla := dbmodels.CLA{
		Text:         this.content,
		OrgSignature: this.OrgSignature,
		CLAInfo:      this.CLAInfo,
	}

	return dbmodels.GetDB().AddCLA(orgRepo, dbmodels.ApplyToCorporation, &cla)
}

func (this *CLACreateOption) AddCLA(orgRepo *dbmodels.OrgRepo) error {
	cla := dbmodels.CLA{
		Text:    this.content,
		CLAInfo: this.CLAInfo,
	}

	return dbmodels.GetDB().AddCLA(orgRepo, dbmodels.ApplyToIndividual, &cla)
}

func (this *CLACreateOption) Validate() (string, error) {
	if len(this.Fields) <= 0 {
		return util.ErrInvalidParameter, fmt.Errorf("no fields")
	}

	if len(this.Fields) > conf.AppConfig.CLAFieldsNumber {
		return util.ErrInvalidParameter, fmt.Errorf("exceeds the max fields number")
	}

	for _, item := range this.Fields {
		if _, err := strconv.Atoi(item.ID); err != nil {
			return util.ErrInvalidParameter, fmt.Errorf("invalid field id")
		}
	}

	text, err := downloadCLA(this.URL)
	if err != nil {
		return util.ErrSystemError, err
	}
	this.content = text

	return "", nil
}

func DeleteCLA(orgRepo *dbmodels.OrgRepo, applyTo, language string) error {
	return dbmodels.GetDB().DeleteCLA(orgRepo, applyTo, language)
}

func GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().GetCLAByType(orgRepo, applyTo)
}

func GetAllCLA(orgRepo *dbmodels.OrgRepo) (*dbmodels.CLAOfLink, error) {
	return dbmodels.GetDB().GetAllCLA(orgRepo)
}

func DownloadOrgSignature(orgRepo *dbmodels.OrgRepo, applyTo string) ([]byte, error) {
	return dbmodels.GetDB().DownloadOrgSignature(orgRepo, applyTo)
}

func DownloadBlankSignature(language string) ([]byte, error) {
	return dbmodels.GetDB().DownloadBlankSignature(language)
}

func downloadCLA(url string) ([]byte, error) {
	var resp *http.Response

	for i := 0; i < 3; i++ {
		v, err := http.Get(url)
		if err == nil {
			resp = v
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	if resp == nil {
		return nil, fmt.Errorf("can't download %s", url)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(http.DetectContentType(data), "text/plain") {
		return data, nil
	}

	return nil, fmt.Errorf("it is not the content of cla")
}
