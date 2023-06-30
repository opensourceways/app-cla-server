package repositoryimpl

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewIndividualSigning(dao dao) *individualSigning {
	return &individualSigning{
		dao: dao,
	}
}

type individualSigning struct {
	dao dao
}

func (impl *individualSigning) Add(is *domain.IndividualSigning) error {
	do := toIndividualSigningDO(is)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	filter := linkIdFilter(is.Link.Id)
	filter[fieldEmail] = is.Rep.EmailAddr.EmailAddr()
	filter[fieldDeleted] = false

	_, err = impl.dao.InsertDocIfNotExists(filter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *individualSigning) Count(linkId string, email dp.EmailAddr) (int, error) {
	filter := linkIdFilter(linkId)
	filter[fieldEmail] = email.EmailAddr()
	filter[fieldDeleted] = false

	var do individualSigningDO

	if err := impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return 0, nil
		}

		return 0, err
	}

	return 1, nil
}
