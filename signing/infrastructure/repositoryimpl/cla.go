package repositoryimpl

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *link) AddCLA(link *domain.Link, cla *domain.CLA) error {
	if err := impl.claContent.add(link.Id, cla); err != nil {
		return err
	}

	do := toCLADO(cla)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.PushArraySingleItemAndUpdate(
		impl.docFilter(link.Id), fieldCLAs, doc,
		bson.M{fieldCLANum: link.CLANum},
		link.Version,
	)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *link) UpdateCLA(link *domain.Link, newCla *domain.CLA) error {
	if err := impl.claContent.add(link.Id, newCla); err != nil {
		return err
	}

	oldCla := link.GetCLA(newCla.Type, newCla.Language)
	if oldCla == nil {
		return commonRepo.NewErrorResourceNotFound(errors.New("can not find old cla"))
	}

	newCla.Fields = oldCla.Fields

	oldDo := toCLADO(oldCla)
	oldDoc, err := oldDo.toDoc()
	if err != nil {
		return err
	}

	newDo := toCLADO(newCla)
	newDoc, err := newDo.toDoc()
	if err != nil {
		return err
	}

	if err = impl.dao.MoveArrayItem(impl.docFilter(link.Id), fieldCLAs,
		bson.M{fieldId: oldCla.Id}, fieldRemoved, oldDoc, link.Version,
	); err != nil {
		return err
	}

	update := bson.M{fieldCLANum: link.CLANum}
	return impl.dao.PushArraySingleItemAndUpdate(impl.docFilter(link.Id), fieldCLAs, newDoc, update, link.Version+1)
}

func (impl *link) RemoveCLA(link *domain.Link, cla *domain.CLA) error {
	do := toCLADO(cla)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.MoveArrayItem(
		impl.docFilter(link.Id), fieldCLAs, bson.M{fieldId: cla.Id},
		fieldRemoved, doc, link.Version,
	)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}
