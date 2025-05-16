package app

import (
	"time"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewIndividualSigningService(
	vc vcservice.VCService,
	repo repository.IndividualSigning,
	corpRepo repository.CorpSigning,
	linkRepo repository.Link,
	interval time.Duration,
) *individualSigningService {
	return &individualSigningService{
		vc:       verificationCodeService{vc},
		repo:     repo,
		corpRepo: corpRepo,
		linkRepo: linkRepo,
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
	repo     repository.IndividualSigning
	corpRepo repository.CorpSigning
	linkRepo repository.Link
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
	claInfo, signed, err := s.SignCheck(cmd)
	if err != nil && !commonRepo.IsErrorResourceNotFound(err) {
		return
	}

	if !signed {
		err = nil
		return
	}

	link, err := s.linkRepo.Find(cmd.LinkId)
	if err != nil {
		return
	}

	logs.Error("link_id is %v,cla info is %v", cmd.LinkId, claInfo)

	isMatch, err := link.AgreementVersionMatch(claInfo)
	if err != nil {
		return
	}

	if isMatch {
		return IndividualSignedDTO{Signed: true}, nil
	} else {
		return IndividualSignedDTO{Signed: false, Reason: "agreement_not_match"}, nil
	}
}

func (s *individualSigningService) SignCheck(cmd *CmdToCheckSinging) (domain.CLAInfo, bool, error) {
	is, err := s.repo.Find(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		logs.Error("find individual sign for %v failed: %v", cmd.EmailAddr.EmailAddr(), err)
	} else {
		return is.Link.CLAInfo, true, nil
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	logs.Error("find employee %v, %v", v, err)
	if err != nil {
		return domain.CLAInfo{}, false, err
	} else {
		info := domain.CLAInfo{CLAId: v.CLAId, AgreementVersion: v.AgreementVersion}

		return info, v.Enabled, nil
	}
}
