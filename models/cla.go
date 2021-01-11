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

type CLA dbmodels.CLA

func (this *CLA) get(onlyFields bool) error {
	v, err := dbmodels.GetDB().GetCLA(this.ID, onlyFields)
	if err != nil {
		return err
	}
	*((*dbmodels.CLA)(this)) = v
	return err
}

func (this *CLA) Get() error {
	return this.get(false)
}

func (this *CLA) GetFields() error {
	return this.get(true)
}

type CLAField = dbmodels.Field

type CLAListOptions dbmodels.CLAListOptions

func (this CLAListOptions) Get() ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().ListCLA(dbmodels.CLAListOptions(this))
}

func ListCLAByIDs(ids []string) ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().ListCLAByIDs(ids)
}

type CLACreateOpt struct {
	dbmodels.CLAData

	orgSignature *[]byte `json:"-"`
	content      *[]byte `json:"-"`
}

func (this *CLACreateOpt) SetOrgSignature(data *[]byte) {
	this.orgSignature = data
}

func (this *CLACreateOpt) toCLACreateOption() *dbmodels.CLACreateOption {
	return &dbmodels.CLACreateOption{
		CLADetail: dbmodels.CLADetail{
			CLAData: this.CLAData,
			Text:    string(*this.content),
			CLAHash: util.Md5sumOfBytes(this.content),
		},
		OrgSignature:     this.orgSignature,
		OrgSignatureHash: util.Md5sumOfBytes(this.orgSignature),
	}
}

func (this *CLACreateOpt) SaveSignatueAtLocal(path string) error {
	if this.orgSignature == nil {
		return nil
	}

	os.Remove(path)
	return ioutil.WriteFile(path, *this.orgSignature, 0644)
}

func (this *CLACreateOpt) SaveCLAAtLocal(path string) error {
	if this.content == nil {
		return nil
	}

	os.Remove(path)
	return ioutil.WriteFile(path, *this.content, 0644)
}

func (this *CLACreateOpt) AddCLA(linkID, applyTo string) IModelError {
	err := dbmodels.GetDB().AddCLA(linkID, applyTo, this.toCLACreateOption())
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrCLAExists, err)
	}

	return parseDBError(err)
}

func (this *CLACreateOpt) AddCLAInfo(linkID, applyTo string) IModelError {
	err := dbmodels.GetDB().AddCLAInfo(linkID, applyTo, this.GenCLAInfo())
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrCLAExists, err)
	}
	return parseDBError(err)
}

func (this *CLACreateOpt) GenCLAInfo() *CLAInfo {
	return &CLAInfo{
		OrgSignatureHash: util.Md5sumOfBytes(this.orgSignature),
		CLAHash:          util.Md5sumOfBytes(this.content),
		CLALang:          this.Language,
		Fields:           this.Fields,
	}
}

func (this *CLACreateOpt) Validate(applyTo string, langs map[string]bool) IModelError {
	this.Language = strings.ToLower(this.Language)

	if applyTo == dbmodels.ApplyToCorporation && !langs[this.Language] {
		return newModelError(ErrUnsupportedCLALang, fmt.Errorf("unsupported_cla_lang"))
	}

	if len(this.Fields) <= 0 {
		return newModelError(ErrNoCLAField, fmt.Errorf("no fields"))
	}

	if len(this.Fields) > conf.AppConfig.CLAFieldsNumber {
		return newModelError(ErrManyCLAField, fmt.Errorf("exceeds the max fields number"))
	}

	for _, item := range this.Fields {
		if _, err := strconv.Atoi(item.ID); err != nil {
			return newModelError(ErrCLAFieldID, fmt.Errorf("invalid field id"))
		}
	}

	text, err := downloadCLA(this.URL)
	if err != nil {
		return newModelError(ErrSystemError, err)
	}
	this.content = text

	if applyTo == dbmodels.ApplyToCorporation && this.orgSignature == nil {
		return newModelError(ErrNoOrgSignature, fmt.Errorf("no signatrue"))
	}

	return nil
}

func downloadCLA(url string) (*[]byte, error) {
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
		return &data, nil
	}

	return nil, fmt.Errorf("it is not the content of cla")
}

func GetCLAByType(orgRepo *dbmodels.OrgRepo, applyTo string) (string, []dbmodels.CLADetail, IModelError) {
	linkID, v, err := dbmodels.GetDB().GetCLAByType(orgRepo, applyTo)
	if err == nil {
		return linkID, v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return linkID, v, newModelError(ErrNoLink, err)
	}
	return linkID, v, parseDBError(err)
}

func GetAllCLA(linkID string) (*dbmodels.CLAOfLink, IModelError) {
	v, err := dbmodels.GetDB().GetAllCLA(linkID)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}

func HasCLA(linkID, applyTo, language string) (bool, IModelError) {
	v, err := dbmodels.GetDB().HasCLA(linkID, applyTo, language)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}

func DeleteCLAInfo(linkID, applyTo, language string) IModelError {
	err := dbmodels.GetDB().DeleteCLAInfo(linkID, applyTo, language)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}

func DeleteCLA(linkID, applyTo, language string) IModelError {
	err := dbmodels.GetDB().DeleteCLA(linkID, applyTo, language)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}

func GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, IModelError) {
	info, err := dbmodels.GetDB().GetCLAInfoSigned(linkID, claLang, applyTo)
	if err == nil {
		return info, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return info, newModelError(ErrNoLinkOrUnsigned, err)
	}
	return info, parseDBError(err)
}

func GetCLAInfoToSign(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, IModelError) {
	v, err := dbmodels.GetDB().GetCLAInfoToSign(linkID, claLang, applyTo)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}
