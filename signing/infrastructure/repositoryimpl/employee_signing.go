package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
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

func (impl *corpSigning) RemoveEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	filter, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	v := toEmployeeSigningDO(es)
	doc, err := v.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.MoveArrayItem(
		filter, fieldEmployees, bson.M{fieldId: es.Id}, fieldDeleted, doc, cs.Version,
	)
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

func (impl *corpSigning) FindEmployees(csId string) ([]domain.EmployeeSigning, error) {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return nil, err
	}

	var do corpSigningDO

	if err = impl.dao.GetDoc(filter, bson.M{fieldEmployees: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return nil, err
	}

	return do.toEmployeeSignings()
}

func (impl *corpSigning) FindEmployeesByEmail(linkId string, email dp.EmailAddr) (
	r repository.EmployeeSigningSummary, err error,
) {
	filter := linkIdFilter(linkId)
	filter[fieldEmployees] = bson.M{mongodbCmdElemMatch: bson.M{fieldEmail: email.EmailAddr()}}

	var do corpSigningDO

	if err = impl.dao.GetDoc(filter, bson.M{fieldEmployees: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return
	}

	v := email.EmailAddr()
	items := do.Employees

	for j := range items {
		if items[j].Email == v {
			r.Enabled = items[j].Enabled

			break
		}
	}

	return
}

func (impl *corpSigning) hasSignedEmployeeCLA(index *domain.CLAIndex) (
	bool, error,
) {
	filter := linkIdFilter(index.LinkId)
	filter[fieldEmployees] = bson.M{mongodbCmdElemMatch: bson.M{fieldCLAId: index.CLAId}}

	var do corpSigningDO

	if err := impl.dao.GetDoc(filter, bson.M{fieldLinkId: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
