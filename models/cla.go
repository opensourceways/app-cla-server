package models

import (
	"io/ioutil"
	"os"

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

type claFileds interface {
	Number() int
	Has(t string) bool
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
