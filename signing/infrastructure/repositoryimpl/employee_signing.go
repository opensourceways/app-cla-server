package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

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

func (impl *corpSigning) SaveEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
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

	err = impl.dao.UpdateArraySingleItem(
		index, fieldEmployees,
		bson.M{fieldId: es.Id}, doc, cs.Version,
	)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}
