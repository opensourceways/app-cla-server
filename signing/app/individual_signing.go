package app

import (
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewIndividualSigningService(
	vc vcservice.VCService,
	repo repository.IndividualSigning,
	corpRepo repository.CorpSigning,
	interval time.Duration,
) *individualSigningService {
	return &individualSigningService{
		vc:       verificationCodeService{vc},
		repo:     repo,
		corpRepo: corpRepo,
		interval: interval,
	}
}

type IndividualSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignIndividualCLA) error
	Check(cmd *CmdToCheckSinging) (bool, error)
}

type individualSigningService struct {
	vc       verificationCodeService
	repo     repository.IndividualSigning
	corpRepo repository.CorpSigning
	interval time.Duration
}

func (s *individualSigningService) Verify(cmd *CmdToCreateVerificationCode) (string, error) {
	return s.vc.newCodeIfItCan((*cmdToCreateCodeForIndividualSigning)(cmd), s.interval)
}

// Sign
func (s *individualSigningService) Sign(cmd *CmdToSignIndividualCLA) error {
	cmd1 := cmd.toCmd()
	if err := s.vc.validate(&cmd1, cmd.VerificationCode); err != nil {
		return err
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
	if cmd.Individual {
		n, err := s.repo.Count(cmd.LinkId, cmd.EmailAddr)
		if err != nil {
			return false, err
		}

		return n > 0, nil
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return v.Enabled, nil
}
