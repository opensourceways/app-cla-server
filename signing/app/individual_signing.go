package app

import (
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewIndividualSigningService(
	vc vcservice.VCService,
	cla claservice.CLAService,
	repo repository.IndividualSigning,
	corpRepo repository.CorpSigning,
	interval time.Duration,
) *individualSigningService {
	return &individualSigningService{
		vc:       verificationCodeService{vc},
		cla:      cla,
		repo:     repo,
		corpRepo: corpRepo,
		interval: interval,
	}
}

type IndividualSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignIndividualCLA) error
	AgreeNewCLA(cmd *CmdToSignIndividualCLA) error
	Check(cmd *CmdToCheckSinging) (IndividualSignedDTO, error)
}

type individualSigningService struct {
	vc       verificationCodeService
	cla      claservice.CLAService
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

	v, err := s.corpRepo.FindCorpSummary(cmd.Link.Id, cmd.Rep.EmailAddr.Domain())
	if err != nil {
		return err
	}
	if len(v) > 0 {
		return domain.NewDomainError(domain.ErrorCodeIndividualSigningCorpExists)
	}

	is := domain.NewIndividualSigning(cmd.Link, cmd.Rep, cmd.AllSingingInfo)
	if err = s.repo.Add(&is); err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeIndividualSigningReSigning)
		}

		return err
	}

	return nil
}

func (s *individualSigningService) AgreeNewCLA(cmd *CmdToSignIndividualCLA) error {
	cmd1 := cmd.toCmd()
	if err := s.vc.validate(&cmd1, cmd.VerificationCode); err != nil {
		return err
	}

	if !s.cla.ContainsCla(cmd.Link.Id, cmd.Link.CLAId) {
		return domain.NewDomainError(domain.ErrorCodeCLANotExists)
	}

	sign, err := s.repo.Find(cmd.Link.Id, cmd.Rep.EmailAddr)
	if err != nil {
		return err
	}

	sign.AgreeNewCLA(cmd.Link.CLAId)

	return s.repo.SaveNewCLA(&sign)
}

// Check
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) (dto IndividualSignedDTO, err error) {
	f := func(claId string, t dp.CLAType) {
		dto.Signed = true
		dto.VersionMatched = s.cla.ContainsCla(cmd.LinkId, claId)
		if !dto.VersionMatched {
			dto.Type = t.CLAType()
		}
	}

	claId, err := s.repo.FindSignedCLA(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		return
	}

	if claId != "" {
		f(claId, dp.CLATypeIndividual)

		return
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = nil
		}

		return dto, err
	}

	if !v.Enabled {
		return dto, nil
	}

	f(v.ClaId, dp.CLATypeCorp)

	return
}
