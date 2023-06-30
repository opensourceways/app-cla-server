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
	Check(cmd *CmdToCheckSinging) (bool, error)
}

type individualSigningService struct {
	repo     repository.IndividualSigning
	corpRepo repository.CorpSigning
}

// Sign
func (s *individualSigningService) Sign(cmd *CmdToSignIndividualCLA) error {
	n, err := s.corpRepo.Count(cmd.Link.Id, cmd.Rep.EmailAddr.Domain())
	if err != nil {
		return err
	}
	if n > 0 {
		return domain.NewDomainError(domain.ErrorCodeIndividualSigningCorpExists)
	}

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
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) (bool, error) {
	n, err := s.repo.Count(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err != nil || len(v) == 0 {
		return false, err
	}

	return v[0].Enabled, nil
}
