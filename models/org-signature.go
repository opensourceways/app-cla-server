package models

import (
	"github.com/zengchen1024/cla-server/dbmodels"
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
