package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CLAInfo = dbmodels.CLAInfo

type CLAField = dbmodels.Field

type CLACreateOpt struct {
	dbmodels.CLAData

	hash    string
	content []byte `json:"-"`
}

func (o *CLACreateOpt) SetCLAContent(data []byte) {
	o.content = data
	o.hash = util.Md5sumOfBytes(data)
}

func (o *CLACreateOpt) GetCLAHash() string {
	return o.hash
}

func (o *CLACreateOpt) toCLACreateOption() *dbmodels.CLACreateOption {
	return &dbmodels.CLACreateOption{
		CLADetail: dbmodels.CLADetail{
			CLAData: o.CLAData,
			CLAHash: o.hash,
		},
	}
}

func (this *CLACreateOpt) SaveCLAAtLocal(path string) error {
	if this.content == nil {
		return nil
	}

	os.Remove(path)
	return ioutil.WriteFile(path, this.content, 0644)
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

func (o *CLACreateOpt) UploadCLAPDF(linkID, applyTo string) IModelError {
	key := dbmodels.CLAPDFIndex{
		LinkID: linkID,
		Apply:  applyTo,
		Lang:   o.Language,
		Hash:   o.hash,
	}
	err := dbmodels.GetDB().UploadCLAPDF(key, o.content)
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
		CLAHash: this.hash,
		CLALang: this.Language,
		Fields:  this.Fields,
	}
}

func (this *CLACreateOpt) Validate(applyTo string, langs map[string]bool) IModelError {
	this.Language = strings.ToLower(this.Language)

	switch applyTo {
	case dbmodels.ApplyToCorporation:
		if langs != nil && !langs[this.Language] {
			return newModelError(ErrUnsupportedCLALang, fmt.Errorf("unsupported_cla_lang"))
		}

		return this.validate(config.AppConfig.CLAConfig.AllowedCorpCLAFields)

	case dbmodels.ApplyToIndividual:
		return this.validate(config.AppConfig.CLAConfig.AllowedIndividualCLAFields)

	default:
		return newModelError(ErrSystemError, fmt.Errorf("unknown cla type"))
	}
}

type claFileds interface {
	Number() int
	Has(t string) bool
}

func (this *CLACreateOpt) validate(fields claFileds) IModelError {
	if len(this.Fields) <= 0 {
		return newModelError(ErrNoCLAField, fmt.Errorf("no fields"))
	}

	if len(this.Fields) > fields.Number() {
		return newModelError(ErrManyCLAField, fmt.Errorf("exceeds the max fields number"))
	}

	for i := range this.Fields {
		item := &this.Fields[i]

		if _, err := strconv.Atoi(item.ID); err != nil {
			return newModelError(ErrCLAFieldID, fmt.Errorf("invalid field id"))
		}

		if !fields.Has(item.Type) {
			return newModelError(ErrCLAFieldID, fmt.Errorf("unknown field"))
		}
	}

	text, err := downloadCLA(this.URL)
	if err != nil {
		return newModelError(ErrSystemError, err)
	}
	this.SetCLAContent(text)

	return nil
}

func downloadCLA(url string) ([]byte, error) {
	return util.DownloadFile(url, "pdf", config.AppConfig.MaxSizeOfCLAContent)
}

func GetCLAByType(linkID, applyTo string) ([]dbmodels.CLADetail, IModelError) {
	v, err := dbmodels.GetDB().GetCLAByType(linkID, applyTo)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
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

func DownloadCorpCLAPDF(linkID, lang string) ([]byte, IModelError) {
	v, err := dbmodels.GetDB().DownloadCorpCLAPDF(linkID, lang)
	return v, parseDBError(err)
}

type CLAPDFIndex = dbmodels.CLAPDFIndex

func DownloadCLAPDF(key CLAPDFIndex) ([]byte, IModelError) {
	v, err := dbmodels.GetDB().DownloadCLAPDF(key)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLinkOrUnuploaed, err)
	}
	return v, parseDBError(err)
}

func DeleteCLAPDF(key CLAPDFIndex) IModelError {
	err := dbmodels.GetDB().DeleteCLAPDF(key)
	return parseDBError(err)
}
