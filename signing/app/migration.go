package app

import (
	"strings"

	"github.com/beego/beego/v2/core/logs"
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/localcla"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

const (
	CorporationMigrationPageSize = 5   // 企业分页大小，可以根据实际情况调整
	IndividualMigrationPageSize  = 100 // 个人分页大小，可以根据实际情况调整
)

type CmdToMigrateCommunity struct {
	UserId       string
	SourceLinkId string
	TargetLinkId string
}

type MigrationService interface {
	MigrateCommunityData(cmd *CmdToMigrateCommunity) error
}

func NewMigrationService(
	linkRepo repository.Link,
	corpRepo repository.CorpSigning,
	individualRepo repository.IndividualSigning,
	userRepo repository.User,
	claService claservice.CLAService,
	localCLA localcla.LocalCLA,
) MigrationService {
	return &migrationService{
		linkRepo:       linkRepo,
		corpRepo:       corpRepo,
		individualRepo: individualRepo,
		userRepo:       userRepo,
		claService:     claService,
		localCLA:       localCLA,
	}
}

type migrationService struct {
	linkRepo       repository.Link
	corpRepo       repository.CorpSigning
	individualRepo repository.IndividualSigning
	userRepo       repository.User
	claService     claservice.CLAService
	localCLA       localcla.LocalCLA
}

func (s *migrationService) MigrateCommunityData(cmd *CmdToMigrateCommunity) error {
	// 验证用户对两个社区都有管理权限
	sourceLink, err := checkIfCommunityManager(cmd.UserId, cmd.SourceLinkId, s.linkRepo)
	if err != nil {
		return err
	}

	targetLink, err := checkIfCommunityManager(cmd.UserId, cmd.TargetLinkId, s.linkRepo)
	if err != nil {
		return err
	}

	// 检查目标社区是否为空
	if err := s.checkTargetLinkIsEmpty(cmd.TargetLinkId); err != nil {
		return err
	}

	// 1. 迁移 CLA 文档
	claIdMap := make(map[string]string) // oldCLAId -> newCLAId
	if err := s.migrateCLADocuments(sourceLink, targetLink, claIdMap); err != nil {
		return err
	}

	corpSigningIdMap := make(map[string]string)

	// 2. 迁移企业签署数据（包含 PDF、email domains、managers）
	if err := s.migrateCorpSigningData(cmd.SourceLinkId, cmd.TargetLinkId, corpSigningIdMap, claIdMap); err != nil {
		return err
	}

	// 3. 迁移个人签署数据
	if err := s.migrateIndividualSigningData(cmd.SourceLinkId, cmd.TargetLinkId, claIdMap); err != nil {
		return err
	}

	// 4. 迁移用户账号信息
	if err := s.migrateUserData(cmd.SourceLinkId, cmd.TargetLinkId, corpSigningIdMap); err != nil {
		return err
	}

	return nil
}

func (s *migrationService) migrateCLADocuments(sourceLink, targetLink *domain.Link, claIdMap map[string]string) error {
	logs.Info("开始迁移 CLA 文档,源社区有 %d 个 CLA,目标社区有 %d 个 CLA",
		len(sourceLink.CLAs), len(targetLink.CLAs))

	if len(sourceLink.CLAs) == 0 {
		logs.Warning("源社区没有 CLA 文档")
		return nil
	}

	// 遍历源社区的所有 CLA
	for _, sourceCLA := range sourceLink.CLAs {
		// 在目标社区中查找匹配的 CLA (Type 和 Language 都相同)
		targetCLA := targetLink.GetCLA(sourceCLA.Type, sourceCLA.Language)

		if targetCLA == nil {
			logs.Warning("目标社区中未找到匹配的 CLA: type=%s, language=%s",
				sourceCLA.Type.CLAType(), sourceCLA.Language.Language())
			continue
		}

		// 建立映射关系
		claIdMap[sourceCLA.Id] = targetCLA.Id

		logs.Info("建立 CLA 映射: %s -> %s (type: %s, lang: %s)",
			sourceCLA.Id, targetCLA.Id, sourceCLA.Type.CLAType(), sourceCLA.Language.Language())
	}

	return nil
}

// func (s *migrationService) migrateCLADocuments(sourceLink, targetLink *domain.Link, claIdMap map[string]string) error {
// 	logs.Info("开始迁移 CLA 文档，源社区有 %d 个 CLA", len(sourceLink.CLAs))
// 	if len(sourceLink.CLAs) == 0 {
// 		logs.Warning("源社区没有 CLA 文档")
// 		return nil
// 	}
// 	// 遍历源社区的所有 CLA
// 	for _, sourceCLA := range sourceLink.CLAs {
// 		// 读取源 CLA 文件内容
// 		sourcePath := s.localCLA.LocalPath(&domain.CLAIndex{
// 			LinkId: sourceLink.Id,
// 			CLAId:  sourceCLA.Id,
// 		})

// 		claText, err := os.ReadFile(sourcePath)
// 		if err != nil {
// 			logs.Error("Failed to read CLA file: %s, err: %v", sourcePath, err)
// 			return err
// 		}

// 		// 创建新的 CLA 对象
// 		newCLA := domain.CLA{
// 			URL:      sourceCLA.URL,
// 			Text:     claText,
// 			Type:     sourceCLA.Type,
// 			Language: sourceCLA.Language,
// 			Fields:   sourceCLA.Fields,
// 		}
// 		// *** 关键修改：每次添加 CLA 前重新获取最新的 targetLink *** 在 MongoDB 更新操作中，系统使用 Version 字段实现乐观锁
// 		freshTargetLink, err := s.linkRepo.Find(targetLink.Id)
// 		if err != nil {
// 			logs.Error("Failed to refresh target link, err: %v", err)
// 			return err
// 		}
// 		// 添加到目标 Link
// 		if err := s.claService.Add(&freshTargetLink, &newCLA); err != nil {
// 			logs.Error("Failed to add CLA to target link, err: %v", err)
// 			return err
// 		}

// 		// 记录 CLA ID 映射关系
// 		claIdMap[sourceCLA.Id] = newCLA.Id

// 		logs.Info("Migrated CLA: %s -> %s (type: %s, lang: %s)",
// 			sourceCLA.Id, newCLA.Id, sourceCLA.Type.CLAType(), sourceCLA.Language.Language())
// 	}

// 	return nil
// }

func (s *migrationService) migrateCorpSigningData(sourceLinkId, targetLinkId string, corpSigningIdMap, claIdMap map[string]string) error {
	totalCount, err := s.corpRepo.CountByLinkId(sourceLinkId)
	if err != nil || totalCount <= 0 {
		return err
	}
	offset := 0
	for int64(offset) < totalCount {
		corpSigningSummaries, err := s.corpRepo.FindAllWithPagination(sourceLinkId, offset, CorporationMigrationPageSize)
		if err != nil {
			return err
		}
		if err := s.migrateBatchCorpSigning(toSummaryInterface(corpSigningSummaries), targetLinkId, claIdMap, corpSigningIdMap); err != nil {
			return err
		}
		currentProgress := offset + len(corpSigningSummaries)
		logs.Info("企业签名迁移进度: %d/%d (%.1f%%)", currentProgress, totalCount, float64(currentProgress)/float64(totalCount)*100)
		offset += CorporationMigrationPageSize
	}
	return nil
}

func toSummaryInterface(summaries []repository.CorpSigningSummary) []corpSigningSummaryAdapter {
	result := make([]corpSigningSummaryAdapter, len(summaries))
	for i := range summaries {
		result[i] = corpSigningSummaryAdapter{
			Id:     summaries[i].Id,
			HasPDF: summaries[i].HasPDF,
		}
	}
	return result
}

func (s *migrationService) migrateBatchCorpSigning(summaries []corpSigningSummaryAdapter, targetLinkId string, claIdMap map[string]string, corpSigningIdMap map[string]string) error {
	for _, summary := range summaries {
		if err := s.migrateOneCorpSigning(summary, targetLinkId, claIdMap, corpSigningIdMap); err != nil {
			return err
		}
	}
	return nil
}

type corpSigningSummaryAdapter struct {
	Id     string
	HasPDF bool
}

func (s *migrationService) migrateOneCorpSigning(summary corpSigningSummaryAdapter, targetLinkId string, claIdMap map[string]string, corpSigningIdMap map[string]string) error {
	fullCorpSigning, err := s.corpRepo.Find(summary.Id)
	if err != nil {
		return err
	}
	oldId := fullCorpSigning.Id
	newCorpSigning := s.cloneCorpSigning(&fullCorpSigning, targetLinkId, claIdMap)
	if err := s.corpRepo.AddForMigrate(&newCorpSigning); err != nil {
		if strings.Contains(err.Error(), "doc exists") || strings.Contains(err.Error(), "document exists") {
			logs.Error("文档已存在错误: 可能重复迁移或ID冲突, oldId=%s, newId=%s", oldId, newCorpSigning.Id)
		}
		return err
	}
	corpSigningIdMap[oldId] = newCorpSigning.Id
	if summary.HasPDF {
		if err := s.migrateCorpPDF(oldId, newCorpSigning.Id); err != nil {
			logs.Error("Failed to migrate PDF for corp signing %s: %v", oldId, err)
		}
	}
	return nil
}

// 迁移企业 PDF
func (s *migrationService) migrateCorpPDF(oldCorpSigningId, newCorpSigningId string) error {
	pdfData, err := s.corpRepo.FindCorpPDF(oldCorpSigningId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil
		}
		if strings.Contains(err.Error(), "cannot decode") {
			logs.Warning("PDF data format issue for corp signing %s, skipping: %v", oldCorpSigningId, err)
			return nil
		}
		return err
	}
	if len(pdfData) == 0 {
		logs.Warning("Empty PDF data for corp signing %s, skipping", oldCorpSigningId)
		return nil
	}
	newCorpSigning, err := s.corpRepo.Find(newCorpSigningId)
	if err != nil {
		return err
	}

	return s.corpRepo.SaveCorpPDF(&newCorpSigning, pdfData)
}

func (s *migrationService) migrateIndividualSigningData(sourceLinkId, targetLinkId string, claIdMap map[string]string) error {
	// 获取总数量
	totalCount, err := s.individualRepo.CountByLinkId(sourceLinkId)
	if err != nil || totalCount <= 0 {
		return err
	}

	// 分页处理
	for offset := 0; int64(offset) < totalCount; offset += IndividualMigrationPageSize {
		individualSignings, err := s.individualRepo.FindAllWithPagination(sourceLinkId, offset, IndividualMigrationPageSize)
		if err != nil {
			return err
		}

		// 批量迁移当前页的数据
		for _, individualSigning := range individualSignings {
			newIndividualSigning := s.cloneIndividualSigning(&individualSigning, targetLinkId, claIdMap)
			if err := s.individualRepo.AddForMigrate(&newIndividualSigning); err != nil {
				// 特别检查是否是文档已存在的错误
				if strings.Contains(err.Error(), "doc exists") || strings.Contains(err.Error(), "document exists") {
					logs.Error("文档已存在错误: 可能重复迁移或ID冲突, newId=%s", individualSigning.Link.Id)
				}
				return err
			}
		}

		// 添加日志记录迁移进度
		logs.Info("Migrated %d/%d individual signings", offset+len(individualSignings), totalCount)
	}

	return nil
}

func (s *migrationService) migrateUserData(sourceLinkId, targetLinkId string, corpSigningIdMap map[string]string) error {
	// 获取源社区的所有用户账号
	users, err := s.userRepo.FindAllByLinkId(sourceLinkId)
	if err != nil {
		return err
	}

	// 为每个用户创建新的账号记录
	for _, user := range users {
		newUser := s.cloneUser(&user, targetLinkId, corpSigningIdMap)
		if _, err := s.userRepo.AddForMigrate(&newUser); err != nil {
			return err
		}
	}

	return nil
}

func (s *migrationService) checkTargetLinkIsEmpty(targetLinkId string) error {
	// 检查是否有企业签署
	hasCorpSigning, err := s.corpRepo.HasSignedLink(targetLinkId)
	if err != nil {
		return err
	}
	if hasCorpSigning {
		return domain.NewDomainError(domain.ErrorCodeLinkCanNotMigrate)
	}

	// 检查是否有个人签署
	hasIndividualSigning, err := s.individualRepo.HasSignedLink(targetLinkId)
	if err != nil {
		return err
	}
	if hasIndividualSigning {
		return domain.NewDomainError(domain.ErrorCodeLinkCanNotMigrate)
	}

	return nil
}

func (s *migrationService) cloneCorpSigning(original *domain.CorpSigning, newLinkId string, claIdMap map[string]string) domain.CorpSigning {
	clone := *original
	clone.Id = ""
	// 更新LinkId到新社区
	clone.Link.Id = newLinkId
	if v, ok := claIdMap[original.Link.CLAId]; ok {
		clone.Link.CLAId = v
	} else {
		logs.Warning("CLA ID mapping not found for %s, setting to empty", original.Link.CLAId)
		clone.Link.CLAId = ""
	}
	return clone
}

func (s *migrationService) cloneIndividualSigning(original *domain.IndividualSigning, newLinkId string, claIdMap map[string]string) domain.IndividualSigning {
	clone := *original
	// clone.Id = ""
	// 更新LinkId到新社区
	clone.Link.Id = newLinkId
	if v, ok := claIdMap[original.Link.CLAId]; ok {
		clone.Link.CLAId = v
	} else {
		logs.Warning("CLA ID mapping not found for %s, setting to empty", original.Link.CLAId)
		clone.Link.CLAId = "0"
	}
	return clone
}

func (s *migrationService) cloneUser(original *domain.User, newLinkId string, corpSigningIdMap map[string]string) domain.User {
	clone := *original
	// 更新LinkId到新社区
	clone.LinkId = newLinkId
	if v, ok := corpSigningIdMap[original.CorpSigningId]; ok {
		clone.CorpSigningId = v
	}
	return clone
}
