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

type OrgCLACreateOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	OrgAlias string `json:"org_alias"`

	ApplyTo string `json:"apply_to"`

	OrgEmail  string `json:"org_email"`
	Submitter string

	CLA CLACreateOption `json:"cla"`
}

func (this OrgCLACreateOption) Validate() (string, error) {
	if this.ApplyTo != dbmodels.ApplyToIndividual && this.ApplyTo != dbmodels.ApplyToCorporation {
		return util.ErrInvalidParameter, fmt.Errorf("invalid apply_to")
	}

	if len(this.CLA.Fields) <= 0 {
		return util.ErrInvalidParameter, fmt.Errorf("no fields")
	}

	if len(this.CLA.Fields) > conf.AppConfig.CLAFieldsNumber {
		return util.ErrInvalidParameter, fmt.Errorf("exceeds the max fields number")
	}

	for _, item := range this.CLA.Fields {
		if _, err := strconv.Atoi(item.ID); err != nil {
			return util.ErrInvalidParameter, fmt.Errorf("invalid field id")
		}
	}

	_, err := dbmodels.GetDB().GetOrgEmailInfo(this.OrgEmail)
	if err == nil {
		return "", nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return util.ErrInvalidEmail, err
	}
	return string(parseDBError(err).ErrCode()), err
}

func (this OrgCLACreateOption) Create(claID string) (string, error) {
	info := dbmodels.OrgCLA{
		Platform:    this.Platform,
		OrgID:       this.OrgID,
		RepoID:      this.RepoID,
		OrgAlias:    this.OrgAlias,
		ApplyTo:     this.ApplyTo,
		OrgEmail:    this.OrgEmail,
		Enabled:     true,
		CLAID:       claID,
		CLALanguage: this.CLA.Language,
		Submitter:   this.Submitter,
	}
	if this.OrgAlias == "" {
		info.OrgAlias = this.OrgID
	}

	return dbmodels.GetDB().CreateOrgCLA(info)
}

type CLACreateOption struct {
	content  []byte
	URL      string           `json:"url"`
	Language string           `json:"language"`
	Fields   []dbmodels.Field `json:"fields"`
}

func (this *CLACreateOption) Create() (string, error) {
	cla := dbmodels.CLA{
		Name:     this.URL,
		Text:     string(this.content),
		Language: this.Language,
		Fields:   this.Fields,
	}
	return dbmodels.GetDB().CreateCLA(cla)
}

func (this *CLACreateOption) Delete(claID string) error {
	return dbmodels.GetDB().DeleteCLA(claID)
}

func (this *CLACreateOption) DownloadCLA() error {
	var resp *http.Response

	for i := 0; i < 3; i++ {
		v, err := http.Get(this.URL)
		if err == nil {
			resp = v
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	if resp == nil {
		return fmt.Errorf("can't download %s", this.URL)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.HasPrefix(http.DetectContentType(data), "text/plain") {
		this.content = data
		return nil
	}

	return fmt.Errorf("it is not the content of cla")
}

type OrgCLA dbmodels.OrgCLA

func (this OrgCLA) Delete() error {
	return dbmodels.GetDB().DeleteOrgCLA(this.ID)
}

func (this *OrgCLA) Get() error {
	v, err := dbmodels.GetDB().GetOrgCLA(this.ID)
	if err != nil {
		return err
	}
	*(*dbmodels.OrgCLA)(this) = v
	return nil
}

type OrgCLAListOption dbmodels.OrgCLAListOption

func (this OrgCLAListOption) List() ([]dbmodels.OrgCLA, error) {
	return dbmodels.GetDB().ListOrgCLA(dbmodels.OrgCLAListOption(this))
}

func ListOrgs(platform string, orgs []string) ([]dbmodels.OrgCLA, error) {
	return dbmodels.GetDB().ListOrgs(platform, orgs)
}

func InitializeIndividualSigning(linkID string) IModelError {
	err := dbmodels.GetDB().InitializeIndividualSigning(linkID)
	return parseDBError(err)
}

type OrgInfo = dbmodels.OrgInfo
type OrgRepo = dbmodels.OrgRepo

func InitializeCorpSigning(linkID string, info *OrgInfo) IModelError {
	err := dbmodels.GetDB().InitializeCorpSigning(linkID, info)
	return parseDBError(err)
}
