package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"time"

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
	Export(userId, linkId string, adminAdded bool, searchQuery string) ([]byte, error)
	FindCorpSummary(cmd *CmdToFindCorpSummary) ([]CorpSummaryDTO, error)
	FindDiffCLAFile(signingId string) (string, error)
	AgreeWithLatestCLA(signingId string) error
	UpdateRepresentative(userId, linkID, signingID, repName, repEmail string) error
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

func (s *corpSigningService) Export(userId, linkId string, adminAdded bool, searchQuery string) ([]byte, error) {
	if _, err := checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return nil, err
	}

	const pageSize = 200
	const maxPages = 500

	var dtos []CorpSigningDTO
	for page := 1; page <= maxPages; page++ {
		v, err := s.repo.FindPage(linkId, page, pageSize, adminAdded, searchQuery)
		if err != nil {
			return nil, err
		}
		if v.Total == 0 {
			break
		}

		for i := range v.Data {
			item := &v.Data[i]
			dtos = append(dtos, CorpSigningDTO{
				Id:             item.Id,
				Date:           item.Date,
				Language:       item.Link.Language.Language(),
				CorpName:       item.Corp.Name.CorpName(),
				RepName:        item.Rep.Name.Name(),
				RepEmail:       item.Rep.EmailAddr.EmailAddr(),
				HasAdminAdded:  !item.Admin.IsEmpty(),
				HasPDFUploaded: item.HasPDF,
			})
		}

		if len(v.Data) < pageSize {
			break
		}
	}

	return corpListToCSV(dtos), nil
}

func corpListToCSV(rows []CorpSigningDTO) []byte {
	var buf bytes.Buffer
	buf.WriteRune('\ufeff')

	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"序号", "企业名称", "申请状态", "CLA语言", "申请时间"})

	for i, row := range rows {
		status := "未完成"
		if row.HasAdminAdded {
			status = "已完成"
		}
		_ = w.Write([]string{
			strconv.Itoa(i + 1),
			escapeCSVCell(row.CorpName),
			status,
			escapeCSVCell(row.Language),
			escapeCSVCell(row.Date),
		})
	}
	w.Flush()

	return buf.Bytes()
}

func escapeCSVCell(s string) string {
	if len(s) == 0 {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@':
		return "'" + s
	}
	return s
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

func (s *corpSigningService) UpdateRepresentative(userId, linkID, signingID, repName, repEmail string) error {
	// 权限验证 - 只有社区管理员可以操作
	if _, err := checkIfCommunityManager(userId, linkID, s.linkRepo); err != nil {
		return err
	}

	// 查找企业签名
	cs, err := s.repo.Find(signingID)
	if err != nil {
		return err
	}

	// 验证link_id匹配
	if cs.Link.Id != linkID {
		return commonRepo.NewErrorResourceNotFound(errors.New("signing not found"))
	}

	// 创建新的代表信息
	newRep, err := domain.NewRepresentative(repName, repEmail)
	if err != nil {
		return err
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
		return err
	}

	// 同步更新 User 表中的邮箱
	if cs.Admin.Id != "" {
		if err := s.userService.UpdateEmail(cs.Link.Id, oldEmail, newRep.EmailAddr); err != nil {
			return err
		}
	}

	return nil
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
