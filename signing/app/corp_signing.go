package app

import (
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewCorpSigningService(
	repo repository.CorpSigning,
	vc vcservice.VCService,
	interval time.Duration,
	linkRepo repository.Link,
	cla claservice.CLAService,
	userService userservice.UserService,
) *corpSigningService {
	return &corpSigningService{
		repo:        repo,
		vc:          verificationCodeService{vc},
		interval:    interval,
		linkRepo:    linkRepo,
		cla:         cla,
		userService: userService,
	}
}

type CorpSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignCorpCLA) error
	Remove(userId, csId string) error
	Get(userId, csId string, email dp.EmailAddr) (string, CorpSigningInfoDTO, error)
	List(userId, linkId string) ([]CorpSigningDTO, error)
	ListPage(userId, linkId string, page, pageSize int, adminAdded bool, searchQuery string) (CorpSigningPageDTO, error)
	FindCorpSummary(cmd *CmdToFindCorpSummary) ([]CorpSummaryDTO, error)
	FindDiffCLAFile(signingId string) (string, error)
	AgreeWithLatestCLA(signingId string) error
	UpdateRepresentative(userId, linkID, signingID, repName, repEmail string) (*ManagerDTO, error)
	FindPendingAgreements(userId, linkId string) ([]CorpSigningPendingDTO, error)
}

type corpSigningService struct {
	vc          verificationCodeService
	cla         claservice.CLAService
	repo        repository.CorpSigning
	interval    time.Duration
	linkRepo    repository.Link
	userService userservice.UserService
}

func (s *corpSigningService) Verify(cmd *CmdToCreateVerificationCode) (string, error) {
	return s.vc.newCodeIfItCan((*cmdToCreateCodeForCorpSigning)(cmd), s.interval)
}

func (s *corpSigningService) Sign(cmd *CmdToSignCorpCLA) error {
	cmd1 := cmd.toCmd()
	if err := s.vc.validate(&cmd1, cmd.VerificationCode); err != nil {
		return err
	}

	v := cmd.toCorpSigning()

	err := s.repo.Add(&v)
	if err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeCorpSigningReSigning)
		}
	}

	return err
}

func (s *corpSigningService) Remove(userId, csId string) error {
	cs, err := s.repo.Find(csId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil
		}

		return err
	}

	if err := cs.CanRemove(); err != nil {
		return err
	}

	if _, err := checkIfCommunityManager(userId, cs.Link.Id, s.linkRepo); err != nil {
		return err
	}

	return s.repo.Remove(&cs)
}

func (s *corpSigningService) Get(userId, csId string, email dp.EmailAddr) (linkId string, dto CorpSigningInfoDTO, err error) {
	item, err := s.repo.Find(csId)
	if err != nil {
		return
	}

	linkId = item.Link.Id

	if !item.IsAdmin(email) {
		if _, err = checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
			return
		}
	}

	dto = CorpSigningInfoDTO{
		Date:     item.Date,
		CLAId:    item.Link.CLAId,
		Language: item.Link.Language.Language(),
		CorpName: item.Corp.Name.CorpName(),
		RepName:  item.Rep.Name.Name(),
		RepEmail: item.Rep.EmailAddr.EmailAddr(),
		AllInfo:  item.AllInfo,
	}

	return
}

func (s *corpSigningService) List(userId, linkId string) ([]CorpSigningDTO, error) {
	if _, err := checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return nil, err
	}

	v, err := s.repo.FindAll(linkId)
	if err != nil || len(v) == 0 {
		return nil, err
	}

	dtos := make([]CorpSigningDTO, len(v))

	for i := range v {
		item := &v[i]

		dtos[i] = CorpSigningDTO{
			Id:             item.Id,
			Date:           item.Date,
			Language:       item.Link.Language.Language(),
			CorpName:       item.Corp.Name.CorpName(),
			RepName:        item.Rep.Name.Name(),
			RepEmail:       item.Rep.EmailAddr.EmailAddr(),
			HasAdminAdded:  !item.Admin.IsEmpty(),
			HasPDFUploaded: item.HasPDF,
		}
	}

	return dtos, nil
}

func (s *corpSigningService) ListPage(userId, linkId string, page, pageSize int, adminAdded bool, searchQuery string) (CorpSigningPageDTO, error) {
	var pageData CorpSigningPageDTO
	pageData.Total = 0
	if _, err := checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return pageData, err
	}

	v, err := s.repo.FindPage(linkId, page, pageSize, adminAdded, searchQuery)
	if err != nil || v.Total == 0 {
		return pageData, err
	}

	dtos := make([]CorpSigningDTO, len(v.Data))

	for i := range v.Data {
		item := &v.Data[i]

		dtos[i] = CorpSigningDTO{
			Id:             item.Id,
			Date:           item.Date,
			Language:       item.Link.Language.Language(),
			CorpName:       item.Corp.Name.CorpName(),
			RepName:        item.Rep.Name.Name(),
			RepEmail:       item.Rep.EmailAddr.EmailAddr(),
			HasAdminAdded:  !item.Admin.IsEmpty(),
			HasPDFUploaded: item.HasPDF,
		}
	}
	pageData.Data = dtos
	pageData.Total = v.Total

	return pageData, nil
}

func (s *corpSigningService) FindCorpSummary(cmd *CmdToFindCorpSummary) ([]CorpSummaryDTO, error) {
	v, err := s.repo.FindCorpSummary(cmd.LinkId, cmd.EmailAddr.Domain())
	if err != nil || len(v) == 0 {
		return nil, err
	}

	r := make([]CorpSummaryDTO, 0, len(v))
	for i := range v {
		if item := &v[i]; item.HasManager {
			r = append(r, CorpSummaryDTO{
				CorpName:      item.CorpName.CorpName(),
				CorpSigningId: item.CorpSigningId,
			})
		}
	}

	return r, nil
}

func (s *corpSigningService) FindDiffCLAFile(signingId string) (string, error) {
	signed, err := s.repo.Find(signingId)
	if err != nil {
		return "", err
	}

	latestClaId := s.cla.GetClaId(signed.Link.Id, dp.CLATypeCorp, signed.Link.Language)
	if latestClaId == "" {
		return "", domain.NewNotFoundDomainError(domain.ErrorCodeCLANotExists)
	}

	if signed.HasSignedCLA(latestClaId) {
		return "", domain.NewDomainError(domain.ErrorCodeCorpSigningCLAIsLatest)
	}

	index := domain.CLAIndex{
		LinkId: signed.Link.Id,
		CLAId:  latestClaId,
	}

	return s.cla.DiffCLALocalFilePath(&index, signed.Link.CLAId)
}

func (s *corpSigningService) AgreeWithLatestCLA(signingId string) error {
	signed, err := s.repo.Find(signingId)
	if err != nil {
		return err
	}

	latestClaId := s.cla.GetClaId(signed.Link.Id, dp.CLATypeCorp, signed.Link.Language)
	if latestClaId == "" {
		return domain.NewNotFoundDomainError(domain.ErrorCodeCLANotExists)
	}

	if err = signed.SetLatestClaId(latestClaId); err != nil {
		return err
	}

	return s.repo.UpdateClaId(&signed)
}

func (s *corpSigningService) UpdateRepresentative(userId, linkID, signingID, repName, repEmail string) (*ManagerDTO, error) {
	// 权限验证 - 只有社区管理员可以操作
	if _, err := checkIfCommunityManager(userId, linkID, s.linkRepo); err != nil {
		return nil, err
	}

	// 查找企业签名
	cs, err := s.repo.Find(signingID)
	if err != nil {
		return nil, err
	}

	// 验证link_id匹配
	if cs.Link.Id != linkID {
		return nil, commonRepo.NewErrorResourceNotFound(errors.New("signing not found"))
	}

	// 创建新的代表信息
	newRep, err := domain.NewRepresentative(repName, repEmail)
	if err != nil {
		return nil, err
	}

	oldEmail := cs.Rep.EmailAddr

	// 更新代表信息
	cs.Rep = newRep

	// 同步更新 Admin 信息（管理员登录账号）
	if cs.Admin.Id != "" {
		cs.Admin.Representative = newRep
	}

	// 保存到数据库
	if err := s.repo.Update(&cs); err != nil {
		return nil, err
	}

	// 无管理员账号或邮箱未变：无需同步 user 表
	if cs.Admin.Id == "" || oldEmail.EmailAddr() == newRep.EmailAddr.EmailAddr() {
		return nil, nil
	}

	// 旧邮箱对应账号已存在：复用并改写邮箱
	exists, err := s.userService.IsAValidUser(cs.Link.Id, oldEmail)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.userService.UpdateEmail(cs.Link.Id, oldEmail, newRep.EmailAddr); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// 旧账号缺失（未预置/迁移遗留）：为当前管理员新建账号
	logs.Info(
		"create user account for the new representative, since the old email(%s) is not found, link_id: %s, signing_id: %s",
		oldEmail.EmailAddr(), cs.Link.Id, signingID,
	)

	pws, _, err := s.userService.Add(cs.Link.Id, signingID, []domain.Manager{cs.Admin})
	if err == nil {
		account, aerr := cs.Admin.Account()
		if aerr != nil {
			return nil, aerr
		}

		admin := &cs.Admin
		return &ManagerDTO{
			Role:      domain.RoleAdmin,
			Name:      admin.Name.Name(),
			Account:   account.Account(),
			Password:  pws[admin.Id].Password(),
			EmailAddr: admin.EmailAddr.EmailAddr(),
		}, nil
	}

	// 新建失败且为账号重复（存在邮箱漂移的孤儿管理员账号）：
	// 按新账号定位孤儿并改写其邮箱，复用已有账号
	if domain.IsErrorOf(err, domain.ErrorCodeUserExists) {
		account, aerr := cs.Admin.Account()
		if aerr != nil {
			return nil, aerr
		}

		logs.Info(
			"fallback to update email by account(%s) for orphan admin, link_id: %s",
			account.Account(), cs.Link.Id,
		)

		if uerr := s.userService.UpdateEmailByAccount(cs.Link.Id, account, newRep.EmailAddr); uerr != nil {
			return nil, uerr
		}
		return nil, nil
	}

	return nil, err
}

func (s *corpSigningService) FindPendingAgreements(userId, linkId string) ([]CorpSigningPendingDTO, error) {
	if _, err := checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return nil, err
	}

	summaries, err := s.repo.FindPendingAgreements(linkId)
	if err != nil {
		return nil, err
	}

	r := make([]CorpSigningPendingDTO, len(summaries))
	for i := range summaries {
		item := &summaries[i]
		r[i] = CorpSigningPendingDTO{
			Id:         item.Id,
			CorpName:   item.Corp.Name.CorpName(),
			AdminEmail: item.Admin.EmailAddr.EmailAddr(),
		}
	}

	return r, nil
}
