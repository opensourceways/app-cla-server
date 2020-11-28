package models

import "github.com/opensourceways/app-cla-server/dbmodels"

func GetOrgCLAWhenSigningAsCorp(orgRepo *dbmodels.OrgRepo, language, signatureMd5 string) (*dbmodels.OrgCLAForSigning, error) {
	return dbmodels.GetDB().GetOrgCLAWhenSigningAsCorp(orgRepo, language, signatureMd5)
}

func GetOrgCLAWhenSigningAsIndividual(orgRepo *dbmodels.OrgRepo, language string) (*dbmodels.OrgCLAForSigning, error) {
	return dbmodels.GetDB().GetOrgCLAWhenSigningAsIndividual(orgRepo, language)
}
