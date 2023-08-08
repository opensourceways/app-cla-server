package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func (impl *corpSigning) SaveCorpPDF(cs *domain.CorpSigning, pdf []byte) error {
	index, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	err = impl.dao.UpdateDoc(
		index, bson.M{fieldPDF: pdf, fieldHasPDF: true, fieldTriggered: true}, cs.Version,
	)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) FindCorpPDF(csId string) ([]byte, error) {
	filter, err := impl.toCorpSigningIndex(csId)
	if err != nil {
		return nil, err
	}

	var do corpSigningDO

	err = impl.dao.GetDoc(filter, bson.M{fieldPDF: 1}, &do)
	if err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return nil, err
	}

	return do.PDF, nil
}
