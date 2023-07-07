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

type CLAPDFIndex = dbmodels.CLAPDFIndex
