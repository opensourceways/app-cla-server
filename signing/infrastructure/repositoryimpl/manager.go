package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *corpSigning) AddAdmin(cs *domain.CorpSigning) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	v := toManagerDO(&cs.Admin)
	doc, err := v.toDoc()
	if err != nil {
		return err
	}

	err = impl.dao.UpdateDoc(index, bson.M{fieldAdmin: doc}, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) AddEmployeeManagers(cs *domain.CorpSigning, ms []domain.Manager) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	docs := make(bson.A, len(ms))
	for i := range ms {
		v := toManagerDO(&ms[i])

		if docs[i], err = v.toDoc(); err != nil {
			return err
		}
	}

	err = impl.dao.PushArrayMultiItems(index, fieldManagers, docs, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) RemoveEmployeeManagers(cs *domain.CorpSigning, ms []string) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	filterOfItem := bson.M{
		fieldId: bson.M{"$in": ms},
	}

	err = impl.dao.PullArrayMultiItems(index, fieldManagers, filterOfItem, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) FindEmployeeManagers(csId string) ([]domain.Manager, error) {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return nil, err
	}

	var do corpSigningDO

	if err = impl.dao.GetDoc(filter, bson.M{fieldManagers: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return nil, err
	}

	return do.toManagers(), nil
}

func (impl *corpSigning) FindCorpManagers(linkId, emailDomain string) ([]domain.Manager, error) {
	filter := linkIdFilter(linkId)
	filter[childField(fieldCorp, fieldDomains)] = bson.M{mongodbCmdIn: bson.A{emailDomain}}

	project := bson.M{
		fieldAdmin:    1,
		fieldManagers: 1,
	}

	var dos []corpSigningDO
	if err := impl.dao.GetDocs(filter, project, &dos); err != nil {
		return nil, err
	}

	r := make([]domain.Manager, 0, len(dos))
	for i := range dos {
		if v := dos[i].allManagers(); len(v) > 0 {
			r = append(r, v...)
		}
	}

	return r, nil
}
