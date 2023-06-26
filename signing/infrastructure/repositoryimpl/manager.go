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
