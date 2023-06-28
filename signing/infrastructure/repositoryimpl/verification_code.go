package repositoryimpl

import (
	"github.com/beego/beego/v2/core/logs"
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/util"
)

func NewVerificationCode(dao dao) *verificationCode {
	return &verificationCode{
		dao: dao,
	}
}

type verificationCode struct {
	dao dao
}

func (impl *verificationCode) Add(code *domain.VerificationCode) error {
	do := toVerificationCodeDO(code)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.DeleteDocs(bson.M{
		fieldExpiry: bson.M{mongodbCmdLt: util.Now()},
	})
	if err != nil {
		logs.Error("remove expired code failed, err:%s", err.Error())
	}

	_, err = impl.dao.InsertDoc(doc)

	return err
}

func (impl *verificationCode) Find(key *domain.VerificationCodeKey) (domain.VerificationCode, error) {
	filter := toVerificationCodeFilter(key)

	var do verificationCodeDO

	if err := impl.dao.GetDocAndDelete(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return domain.VerificationCode{}, err
	}

	return do.toVerificationCode()
}
