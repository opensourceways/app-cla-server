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

	sign, err := s.repo.Find(cmd.Link.Id, cmd.Rep.EmailAddr)
	if err != nil {
		return err
	}

	latestClaId := s.cla.GetClaId(sign.Link.Id, dp.CLATypeIndividual, sign.Link.Language)
	if latestClaId == "" {
		return domain.NewDomainError(domain.ErrorCodeCLANotExists)
	}

	if err = sign.AgreeNewCLA(latestClaId); err != nil {
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
	claId, language, err := s.repo.FindSignedCLA(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		return
	}

	if claId != "" {
		dto.Signed = true
		dto.Type = dp.CLATypeIndividual.CLAType()
		versionMatched := s.cla.ContainsCla(cmd.LinkId, claId)

		// version_matched 表达：签署是否当前有效（考虑宽限期）
		if versionMatched || s.isInGracePeriod(cmd.LinkId, dp.CLATypeIndividual, language) {
			dto.Status = "valid"
			dto.VersionMatched = true
		} else {
			dto.Status = "expired"
			dto.VersionMatched = false
		}

		// 获取调试信息
		dto.DebugInfo = s.getDebugInfo(cmd.LinkId, dp.CLATypeIndividual, language, versionMatched)

		return
	}

	v, err := s.corpRepo.FindEmployeesByEmail(cmd.LinkId, cmd.EmailAddr)
	if err != nil {
		if !commonRepo.IsErrorResourceNotFound(err) {
			return dto, err
		}
	} else if v.Enabled {
		dto.Signed = true
		dto.Type = dp.CLATypeCorp.CLAType()
		versionMatched := s.cla.ContainsCla(cmd.LinkId, v.ClaId)

		if versionMatched || s.isInGracePeriod(cmd.LinkId, dp.CLATypeCorp, v.Language) {
			dto.Status = "valid"
			dto.VersionMatched = true
		} else {
			dto.Status = "expired"
			dto.VersionMatched = false
		}

		dto.DebugInfo = s.getDebugInfo(cmd.LinkId, dp.CLATypeCorp, v.Language, versionMatched)
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
	dto.Signed = false
	dto.VersionMatched = false

	return
}

// getDebugInfo 获取调试信息
func (s *individualSigningService) getDebugInfo(
	linkId string, claType dp.CLAType, language dp.Language, isLatestClaVersion bool,
) *DebugInfoDTO {
	link, err := s.linkRepo.Find(linkId)
	if err != nil {
		return nil
	}

	lastUpdateTime := s.cla.GetLastUpdateTime(linkId, claType, language)
	if lastUpdateTime.IsZero() {
		return nil
	}

	gracePeriodDays := link.GetEffectiveGracePeriodDays(s.defaultGracePeriodDays)
	elapsedDays := int(time.Since(lastUpdateTime).Hours() / 24)
	gracePeriodEndsAt := lastUpdateTime.AddDate(0, 0, gracePeriodDays)

	return &DebugInfoDTO{
		IsLatestClaVersion:  isLatestClaVersion,
		InGracePeriod:       inGracePeriod(gracePeriodDays, lastUpdateTime),
		GracePeriodDays:     gracePeriodDays,
		ClaUpdatedAt:        lastUpdateTime.Format("2006-01-02"),
		DaysSinceLastUpdate: elapsedDays,
		GracePeriodEndsAt:   gracePeriodEndsAt.Format("2006-01-02"),
	}
}

func (s *individualSigningService) isInGracePeriod(linkId string, claType dp.CLAType, language dp.Language) bool {
	link, err := s.linkRepo.Find(linkId)
	if err != nil {
		return true
	}

	days := link.GetEffectiveGracePeriodDays(s.defaultGracePeriodDays)

	lastUpdateTime := s.cla.GetLastUpdateTime(linkId, claType, language)
	if lastUpdateTime.IsZero() {
		return true
	}

	return inGracePeriod(days, lastUpdateTime)
}

// inGracePeriod 是 isInGracePeriod 与 getDebugInfo 共用的唯一判定口径，避免两处逻辑各算一套、互相矛盾。
// days <= 0 表示没有宽限期，一旦版本不匹配立即视为过期。
func inGracePeriod(days int, lastUpdateTime time.Time) bool {
	return time.Since(lastUpdateTime) < time.Duration(days)*24*time.Hour
}
