package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
)

func GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, *ModelError) {
	info, err := dbmodels.GetDB().GetCLAInfoSigned(linkID, claLang, applyTo)
	return info, parseDBError(err)
}

func DeleteCLAInfo(linkID, applyTo, language string) error {
	return nil
}

func InitializeIndividualSigning(linkID string, orgRepo *dbmodels.OrgRepo, claInfo *dbmodels.CLAInfo) *ModelError {
	err := dbmodels.GetDB().InitializeIndividualSigning(linkID, orgRepo, claInfo)
	return parseDBError(err)
}

func InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo, claInfo *dbmodels.CLAInfo) *ModelError {
	err := dbmodels.GetDB().InitializeCorpSigning(linkID, info, claInfo)
	return parseDBError(err)
}
