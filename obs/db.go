package obs

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type FileStorage struct {
	OBS OBS
}

func (fs FileStorage) UploadCorporationSigningPDF(linkID, adminEmail string, pdf []byte) dbmodels.IDBError {
	err := fs.OBS.WriteObject(buildCorpSigningPDFPath(linkID, adminEmail), pdf)
	return toDBError(err)
}

func (fs FileStorage) DownloadCorporationSigningPDF(linkID, email, path string) dbmodels.IDBError {
	err := fs.OBS.ReadObject(buildCorpSigningPDFPath(linkID, email), path)
	if err.IsObjectNotFound() {
		return dbmodels.NewDBError(dbmodels.ErrNoDBRecord, err)
	}
	return toDBError(err)
}

func (fs FileStorage) IsCorporationSigningPDFUploaded(linkID, email string) (bool, dbmodels.IDBError) {
	b, err := fs.OBS.HasObject(buildCorpSigningPDFPath(linkID, email))
	return b, toDBError(err)
}

func buildCorpSigningPDFPath(linkID string, email string) string {
	return fmt.Sprintf("%s/%s", linkID, util.EmailSuffix(email))
}

func toDBError(err error) dbmodels.IDBError {
	if err == nil {
		return nil
	}
	return dbmodels.NewDBError(dbmodels.ErrSystemError, err)
}
