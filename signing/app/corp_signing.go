package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCorpSigningService(repo repository.CorpSigning) *corpSigningService {
	return &corpSigningService{repo}
}

type CorpSigningService interface {
	Sign(cmd *CmdToSignCorpCLA) error
}

type corpSigningService struct {
	repo repository.CorpSigning
}

func (s *corpSigningService) Sign(cmd *CmdToSignCorpCLA) error {
	v := cmd.toCorpSigning()

	err := s.repo.Add(&v)
	if err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeCorpSigningReSigning)
		}
	}

	return err
}
