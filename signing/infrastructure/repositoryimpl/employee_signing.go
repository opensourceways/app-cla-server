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
	[]repository.EmployeeSigningSummary, error,
) {
	filter := linkIdFilter(linkId)
	filter[childField(fieldCorp, fieldDomains)] = bson.M{mongodbCmdIn: bson.A{email.Domain()}}

	var dos []corpSigningDO

	err := impl.dao.GetArrayItem(
		filter, fieldEmployees,
		bson.M{fieldEmail: email.EmailAddr()},
		bson.M{childField(fieldEmployees, fieldEnabled): 1}, &dos,
	)
	if err != nil {
		return nil, err
	}

	r := make([]repository.EmployeeSigningSummary, 0, len(dos))

	for i := range dos {
		if len(dos[i].Employees) == 0 {
			continue
		}

		r = append(r, repository.EmployeeSigningSummary{
			Enabled: dos[i].Employees[0].Enabled,
		})
	}

	return r, nil
}

func (impl *corpSigning) hasSignedEmployeeCLA(index *domain.CLAIndex) (
	bool, error,
) {
	filter := linkIdFilter(index.LinkId)

	var dos []corpSigningDO

	err := impl.dao.GetArrayItem(
		filter, fieldEmployees,
		bson.M{fieldCLAId: index.CLAId},
		bson.M{childField(fieldEmployees, fieldCLAId): 1}, &dos,
	)
	if err != nil || len(dos) == 0 {
		return false, err
	}

	for i := range dos {
		if len(dos[i].Employees) > 0 {
			return true, nil
		}
	}

	return false, nil
}
