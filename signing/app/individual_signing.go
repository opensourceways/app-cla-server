package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewIndividualSigningService(
	repo repository.IndividualSigning,
	corpRepo repository.CorpSigning,
) *individualSigningService {
	return &individualSigningService{
		repo:     repo,
		corpRepo: corpRepo,
	}
}

type IndividualSigningService interface {
	Sign(cmd *CmdToSignIndividualCLA) error
	Check(cmd *CmdToCheckSinging) bool
}

type individualSigningService struct {
	repo     repository.IndividualSigning
	corpRepo repository.CorpSigning
}

// Sign
func (s *individualSigningService) Sign(cmd *CmdToSignIndividualCLA) error {
	is := cmd.toIndividualSigning()

	if err := s.repo.Add(&is); err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeIndividualSigningReSigning)
		}

		return err
	}

	return nil
}

// Check
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) bool {
	if n, err := s.repo.Count(cmd.LinkId, cmd.EmailAddr); err == nil && n > 0 {
		return true
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err == nil && len(v) > 0 {
		return v[0].Enabled
	}

	return false
}
