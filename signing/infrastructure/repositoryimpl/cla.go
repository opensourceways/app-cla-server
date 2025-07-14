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

	filter := bson.M{
		fieldId:                        link.Id,
		childField(fieldCLAs, fieldId): oldCla.Id,
	}

	update := bson.M{
		fieldCLANum:  link.CLANum,
		"clas.$.id":  newCla.Id,
		"clas.$.url": newCla.URL,
	}
	return impl.dao.PushArraySingleItemAndUpdate(filter, fieldRemoved, oldDoc, update, link.Version)
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
