package repositoryimpl

import (
	//"go.mongodb.org/mongo-driver/bson"

	//commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
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

func (impl *emailCredential) Add(*domain.EmailCredential) error {
	return nil
}
func (impl *emailCredential) Find(dp.EmailAddr) (domain.EmailCredential, error) {
	return domain.EmailCredential{}, nil
}
