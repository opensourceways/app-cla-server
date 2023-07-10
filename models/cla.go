package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CLAInfo = dbmodels.CLAInfo

type CLAField = dbmodels.Field

type CLACreateOpt struct {
	dbmodels.CLAData

	hash    string
	content []byte `json:"-"`
}

type CLAPDFIndex = dbmodels.CLAPDFIndex
