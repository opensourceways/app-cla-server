package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldCode    = "code"
	fieldExpiry  = "expiry"
	fieldPurpose = "purpose"
)

func toVerificationCodeDO(code *domain.VerificationCode) verificationCodeDO {
	return verificationCodeDO{
		Code:    code.Code,
		Expiry:  code.Expiry,
		Purpose: code.Purpose.Purpose(),
	}
}

func toVerificationCodeFilter(key *domain.VerificationCodeKey) bson.M {
	return bson.M{
		fieldCode:    key.Code,
		fieldPurpose: key.Purpose.Purpose(),
	}
}

type verificationCodeDO struct {
	Code    string `bson:"code"     json:"code"     required:"true"`
	Expiry  int64  `bson:"expiry"   json:"expiry"   required:"true"`
	Purpose string `bson:"purpose"  json:"purpose"  required:"true"`
}

func (do *verificationCodeDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *verificationCodeDO) toVerificationCode() (r domain.VerificationCode, err error) {
	if r.Purpose, err = dp.NewPurpose(do.Purpose); err != nil {
		return
	}

	r.Code = do.Code
	r.Expiry = do.Expiry

	return
}
