package app

import (
	"time"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
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
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) (dto IndividualSignedDTO, err error) {
	f := func(info domain.LinkInfo) (dto IndividualSignedDTO, err error) {
		isValid := s.cla.CheckCla(info)
		if isValid {
			return IndividualSignedDTO{Signed: true}, nil
		} else {
			return IndividualSignedDTO{Signed: false, Reason: "agreement_is_outdated"}, nil
		}
	}

	is, err := s.repo.Find(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		if !commonRepo.IsErrorResourceNotFound(err) {
			logs.Error("find individual sign for %v failed: %v", cmd.EmailAddr.EmailAddr(), err)
		}
	} else {
		return f(is.Link)
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

	return f(v.Link)
}
