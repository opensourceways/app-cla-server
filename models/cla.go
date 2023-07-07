package models

import (
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

type CLAPDFIndex = dbmodels.CLAPDFIndex
