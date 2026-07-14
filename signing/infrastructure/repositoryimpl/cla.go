package repositoryimpl

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *link) AddCLA(link *domain.Link, cla *domain.CLA) error {
	if err := impl.claContent.add(link.Id, cla); err != nil {
		return err
	}

	cla.UpdatedAt = time.Now().Unix()

	do := toCLADO(cla)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.PushArraySingleItemAndUpdate(
		impl.docFilter(link.Id),
		fieldCLAs,
		doc,
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

	oldDo := toCLADO(oldCla)
	oldDoc, err := oldDo.toDoc()
	if err != nil {
		return err
	}

	filter := bson.M{
		fieldId: link.Id,
	}

	filterOfArray := bson.M{
		fieldLang: oldCla.Language.Language(),
		fieldType: oldCla.Type.CLAType(),
	}

	newCla.UpdatedAt = time.Now().Unix()

	update := bson.M{
		fieldCLANum: link.CLANum,
		fieldClaUpdatedAt: newCla.UpdatedAt,
	}

	otherSet := bson.M{
		fieldCLANum: link.CLANum,
	}

	return impl.dao.PushAndUpdateArrayItem(
		filter, fieldRemoved, oldDoc, fieldCLAs, filterOfArray, update, link.Version, otherSet,
	)
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
