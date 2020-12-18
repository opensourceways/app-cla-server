package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
)

func GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, *ModelError) {
	info, err := dbmodels.GetDB().GetCLAInfoSigned(linkID, claLang, applyTo)
	return info, parseDBError(err)
}

func AddCLAInfo(linkID, applyTo string, info *dbmodels.CLAInfo) error {
	return dbmodels.GetDB().AddCLAInfo(linkID, applyTo, info)
}

func DeleteCLAInfo(linkID, applyTo, language string) error {
	return nil
}
