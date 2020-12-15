package models

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CLACreateOption struct {
	dbmodels.CLAData

	OrgSignature []byte `json:"-"`
	content      []byte `json:"-"`
}

func (this *CLACreateOption) AddCLA(linkID string) error {
	cla := dbmodels.CLA{
		Text:         this.content,
		CLAData:      this.CLAData,
		OrgSignature: this.OrgSignature,
	}

	return dbmodels.GetDB().AddCLA(linkID, dbmodels.ApplyToIndividual, &cla)
}

func (this *CLACreateOption) AddCLAInfo(linkID string) error {
	return dbmodels.GetDB().AddCLAInfo(linkID, dbmodels.ApplyToIndividual, this.GenCLAInfo())
}

func (this *CLACreateOption) GenCLAInfo() *dbmodels.CLAInfo {
	return &dbmodels.CLAInfo{
		OrgSignatureHash: util.Md5sumOfBytes(this.OrgSignature),
		CLAHash:          util.Md5sumOfBytes(this.content),
		CLALang:          this.Language,
		Fields:           this.Fields,
	}
}

func (this *CLACreateOption) SaveSignatueAtLocal(path string) error {
	if this.OrgSignature == nil {
		return nil
	}

	os.Remove(path)
	return ioutil.WriteFile(path, this.OrgSignature, 0644)
}

func (this *CLACreateOption) SaveCLAAtLocal(path string) error {
	if this.content == nil {
		return nil
	}

	os.Remove(path)
	return ioutil.WriteFile(path, this.content, 0644)
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

func DeleteCLA(linkID, applyTo, language string) error {
	return dbmodels.GetDB().DeleteCLA(linkID, applyTo, language)
}

func GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().GetCLAByType(orgRepo, applyTo)
}

func GetAllCLA(linkID string) (*dbmodels.CLAOfLink, error) {
	return dbmodels.GetDB().GetAllCLA(linkID)
}

func HasCLA(linkID, applyTo, language string) (bool, error) {
	return dbmodels.GetDB().HasCLA(linkID, applyTo, language)
}

func DownloadOrgSignature(linkID, language string) ([]byte, error) {
	// return dbmodels.GetDB().DownloadOrgSignature(orgRepo, language)
	return dbmodels.GetDB().DownloadOrgSignature(language)
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
