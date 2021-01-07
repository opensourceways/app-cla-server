package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
)

func UploadOrgSignature(orgCLAID string, pdf []byte) error {
	return dbmodels.GetDB().UploadOrgSignature(orgCLAID, pdf)
}

func DownloadOrgSignature(orgCLAID string) ([]byte, error) {
	return dbmodels.GetDB().DownloadOrgSignature(orgCLAID)
}

func DownloadOrgSignatureByMd5(orgCLAID, md5sum string) ([]byte, error) {
	return dbmodels.GetDB().DownloadOrgSignatureByMd5(orgCLAID, md5sum)
}
