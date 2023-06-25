package repositoryimpl

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *corpSigning) AddEmployee(cs *domain.CorpSigning) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	e := cs.NewestEmployee()
	e.Id = impl.dao.NewDocId()

	v := toEmployeeSigningDO(e)
	doc, err := v.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.PushArrayDoc(index, doc, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}
