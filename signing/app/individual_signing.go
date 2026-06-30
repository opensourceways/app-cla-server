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
	linkRepo repository.Link,
	interval time.Duration,
	defaultGracePeriodDays int,
) *individualSigningService {
	return &individualSigningService{
		vc:                    verificationCodeService{vc},
		cla:                   cla,
		repo:                  repo,
		corpRepo:              corpRepo,
		linkRepo:              linkRepo,
		interval:              interval,
		defaultGracePeriodDays: defaultGracePeriodDays,
	}
}

type IndividualSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignIndividualCLA) error
	AgreeNewCLA(cmd *CmdToSignIndividualCLA) error
	FindDiffCLAFile(cmd *CmdToFindSignedCLAInfo) (string, error)
	Check(cmd *CmdToCheckSinging) (IndividualSignedDTO, error)
}

type individualSigningService struct {
	vc                    verificationCodeService
	cla                   claservice.CLAService
	repo                  repository.IndividualSigning
	corpRepo              repository.CorpSigning
	linkRepo              repository.Link
	interval              time.Duration
	defaultGracePeriodDays int
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

	if err = sign.AgreeNewCLA(cmd.Link.CLAId); err != nil {
		return err
	}

	return s.repo.SaveNewCLA(&sign)
}

func (s *individualSigningService) FindDiffCLAFile(cmd *CmdToFindSignedCLAInfo) (string, error) {
	signed, err := s.repo.Find(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		return "", err
	}

	newClaId := s.cla.GetClaId(signed.Link.Id, dp.CLATypeIndividual, signed.Link.Language)
	if newClaId == "" {
		return "", domain.NewNotFoundDomainError(domain.ErrorCodeCLANotExists)
	}

	if signed.HasSignedCLA(newClaId) {
		return "", domain.NewDomainError(domain.ErrorCodeIndividualSigningCLAIsLatest)
	}

	index := domain.CLAIndex{
		LinkId: cmd.LinkId,
		CLAId:  newClaId,
	}

	return s.cla.DiffCLALocalFilePath(&index, signed.Link.CLAId)
}

// Check
func (s *individualSigningService) Check(cmd *CmdToCheckSinging) (dto IndividualSignedDTO, err error) {
	claId, err := s.repo.FindSignedCLA(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		return
	}

	if claId != "" {
		dto.Type = dp.CLATypeIndividual.CLAType()
		versionMatched := s.cla.ContainsCla(cmd.LinkId, claId)

		if versionMatched {
			// 版本匹配，签署有效
			dto.Status = "valid"
		} else if s.isInGracePeriod(cmd.LinkId) {
			// 版本不匹配但在宽限期内，签署仍有效
			dto.Status = "valid"
		} else {
			// 版本不匹配且超过宽限期，签署已过期
			dto.Status = "expired"
		}

		return
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		if !commonRepo.IsErrorResourceNotFound(err) {
			return dto, err
		}
	} else if v.Enabled {
		dto.Type = dp.CLATypeCorp.CLAType()
		versionMatched := s.cla.ContainsCla(cmd.LinkId, v.ClaId)

		if versionMatched {
			// 版本匹配，签署有效
			dto.Status = "valid"
		} else if s.isInGracePeriod(cmd.LinkId) {
			// 版本不匹配但在宽限期内，签署仍有效
			dto.Status = "valid"
		} else {
			// 版本不匹配且超过宽限期，签署已过期
			dto.Status = "expired"
		}
		return
	}

	// 未签署过
	corps, err := s.corpRepo.FindCorpSummary(cmd.LinkId, cmd.EmailAddr.Domain())
	if err != nil {
		return dto, err
	}

	if len(corps) > 0 {
		dto.Type = dp.CLATypeCorp.CLAType()
	} else {
		dto.Type = dp.CLATypeIndividual.CLAType()
	}

	dto.Status = "not_signed"

	return
}

func (s *individualSigningService) isInGracePeriod(linkId string) bool {
	link, err := s.linkRepo.Find(linkId)
	if err != nil {
		return true
	}

	days := link.GetEffectiveGracePeriodDays(s.defaultGracePeriodDays)
	if days == 0 {
		return true
	}

	lastUpdateTime := s.cla.GetLastUpdateTime(linkId)
	if lastUpdateTime.IsZero() {
		return true
	}

	return time.Since(lastUpdateTime) < time.Duration(days)*24*time.Hour
}
