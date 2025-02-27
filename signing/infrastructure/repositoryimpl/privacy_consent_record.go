package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/util"
)

func NewPrivacyConsentRecord(dao dao) *privacyConsentRecord {
	return &privacyConsentRecord{
		dao: dao,
	}
}

type privacyConsentRecord struct {
	dao dao
}

func (impl *privacyConsentRecord) Add(account, platform, ver string) error {
	do := privacyConsentRecordDO{
		Account:  account,
		Platform: platform,
		PrivacyConsentDO: PrivacyConsentDO{
			Time:    util.Time(),
			Version: ver,
		},
	}
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	index := bson.M{
		fieldAccount:  account,
		fieldPlatform: platform,
	}

	_, err = impl.dao.ReplaceDoc(index, doc)

	return err
}

type PrivacyConsentDO = privacyConsentDO

type privacyConsentRecordDO struct {
	Account  string `bson:"account"   json:"account"   required:"true"`
	Platform string `bson:"platform"  json:"platform"  required:"true"`

	PrivacyConsentDO `bson:",inline"`
}

func (do *privacyConsentRecordDO) toDoc() (bson.M, error) {
	return genDoc(do)
}
