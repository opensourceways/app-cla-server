package repositoryimpl

import (
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

	err = impl.dao.UpdateDoc(index, fieldAdmin, doc, cs.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}
