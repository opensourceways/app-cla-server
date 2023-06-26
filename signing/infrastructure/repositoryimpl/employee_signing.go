package repositoryimpl

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *corpSigning) AddEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	es.Id = impl.dao.NewDocId()
	v := toEmployeeSigningDO(es)
	doc, err := v.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.PushArraySingleItem(index, fieldEmployees, doc, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}
