package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
)

func UploadOrgSignature(claOrgID string, pdf []byte) error {
	return dbmodels.GetDB().UploadOrgSignature(claOrgID, pdf)
}

func DownloadOrgSignature(claOrgID string) ([]byte, error) {
	return dbmodels.GetDB().DownloadOrgSignature(claOrgID)
}

func DownloadBlankSignature(language string) ([]byte, error) {
	return DownloadBlankSignature(language)
}
