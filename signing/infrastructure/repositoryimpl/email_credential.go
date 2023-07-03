package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewEmailCredential(dao dao) *emailCredential {
	return &emailCredential{
		dao: dao,
	}
}

type emailCredential struct {
	dao dao
}

func (impl *emailCredential) Add(ec *domain.EmailCredential) error {
	do := toEmailCredentialDO(ec)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldToken] = ec.Token

	filter := bson.M{fieldEmail: ec.Addr.EmailAddr()}

	_, err = impl.dao.ReplaceDoc(filter, doc)

	return err
}

func (impl *emailCredential) Find(addr dp.EmailAddr) (domain.EmailCredential, error) {
	filter := bson.M{fieldEmail: addr.EmailAddr()}

	var do emailCredentialDO

	if err := impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return domain.EmailCredential{}, err
	}

	return do.toEmailCredential()
}
