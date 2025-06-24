package repositoryimpl

import (
	"fmt"

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

func (impl *link) UpdateCLA(link *domain.Link, oldCla, newCla *domain.CLA) error {
	if err := impl.claContent.add(link.Id, newCla); err != nil {
		return err
	}

	filter := bson.M{
		fieldId:                        link.Id,
		childField(fieldCLAs, fieldId): oldCla.Id,
	}

	do := toCLADO(oldCla)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}

	update := bson.M{
		fieldCLANum: link.CLANum,
		fmt.Sprintf("%s.$.%s", fieldCLAs, fieldId):  newCla.Id,
		fmt.Sprintf("%s.$.%s", fieldCLAs, fieldUrl): newCla.URL,
	}

	return impl.dao.PushArraySingleItemAndUpdate(filter, fieldRemoved, doc, update, link.Version)
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
