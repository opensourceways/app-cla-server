package app

import (
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/opensourceways/app-cla-server/signing/domain"
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
) MigrationService {
	return &migrationService{
		linkRepo:       linkRepo,
		corpRepo:       corpRepo,
		individualRepo: individualRepo,
		userRepo:       userRepo,
	}
}

type migrationService struct {
	linkRepo       repository.Link
	corpRepo       repository.CorpSigning
	individualRepo repository.IndividualSigning
	userRepo       repository.User
}

func (s *migrationService) MigrateCommunityData(cmd *CmdToMigrateCommunity) error {
	// 验证用户对两个社区都有管理权限
	_, err := checkIfCommunityManager(cmd.UserId, cmd.SourceLinkId, s.linkRepo)
	if err != nil {
		return err
	}

	_, err = checkIfCommunityManager(cmd.UserId, cmd.TargetLinkId, s.linkRepo)
	if err != nil {
		return err
	}

	// 检查目标社区是否为空
	if err := s.checkTargetLinkIsEmpty(cmd.TargetLinkId); err != nil {
		return err
	}

	corpSigningIdMap := make(map[string]string)

	// 1. 迁移企业签署数据
	if err := s.migrateCorpSigningData(cmd.SourceLinkId, cmd.TargetLinkId, corpSigningIdMap); err != nil {
		return err
	}

	// 2. 迁移个人签署数据
	if err := s.migrateIndividualSigningData(cmd.SourceLinkId, cmd.TargetLinkId); err != nil {
		return err
	}

	// 3. 新增：迁移用户账号信息
	if err := s.migrateUserData(cmd.SourceLinkId, cmd.TargetLinkId, corpSigningIdMap); err != nil {
		return err
	}

	return nil
}

func (s *migrationService) migrateCorpSigningData(sourceLinkId, targetLinkId string, corpSigningIdMap map[string]string) error {
	// 获取总数量
	totalCount, err := s.corpRepo.CountByLinkId(sourceLinkId)
	if err != nil || totalCount <= 0 {
		return err
	}

	// 分页处理
	for offset := 0; int64(offset) < totalCount; offset += CorporationMigrationPageSize {
		corpSigningSummaries, err := s.corpRepo.FindAllWithPagination(sourceLinkId, offset, CorporationMigrationPageSize)
		if err != nil {
			return err
		}
		// 批量迁移当前页的数据
		for _, summary := range corpSigningSummaries {
			// 使用Find方法获取完整的CorpSigning对象
			fullCorpSigning, err := s.corpRepo.Find(summary.Id)
			if err != nil {
				return err
			}
			oldId := fullCorpSigning.Id
			// 克隆并修改LinkId
			newCorpSigning := s.cloneCorpSigning(&fullCorpSigning, targetLinkId)
			// 添加到新社区并修改CorpSigning.Id
			if err := s.corpRepo.AddForMigrate(&newCorpSigning); err != nil {
				// 特别检查是否是文档已存在的错误
				if strings.Contains(err.Error(), "doc exists") || strings.Contains(err.Error(), "document exists") {
					logs.Error("文档已存在错误: 可能重复迁移或ID冲突, oldId=%s, newId=%s", oldId, newCorpSigning.Id)
				}
				return err
			}
			corpSigningIdMap[oldId] = newCorpSigning.Id
		}
		// 添加日志记录迁移进度
		currentProgress := offset + len(corpSigningSummaries)
		logs.Info("企业签名迁移进度: %d/%d (%.1f%%)", currentProgress, totalCount, float64(currentProgress)/float64(totalCount)*100)
	}
	return nil
}

func (s *migrationService) migrateIndividualSigningData(sourceLinkId, targetLinkId string) error {
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
			newIndividualSigning := s.cloneIndividualSigning(&individualSigning, targetLinkId)
			if err := s.individualRepo.Add(&newIndividualSigning); err != nil {
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

func (s *migrationService) cloneCorpSigning(original *domain.CorpSigning, newLinkId string) domain.CorpSigning {
	clone := *original
	// 更新LinkId到新社区
	clone.Link.Id = newLinkId
	clone.Link.CLAId = "0" // 清空CLAId，避免引用旧社区的CLA
	return clone
}

func (s *migrationService) cloneIndividualSigning(original *domain.IndividualSigning, newLinkId string) domain.IndividualSigning {
	clone := *original
	// 更新LinkId到新社区
	clone.Link.Id = newLinkId
	clone.Link.CLAId = "0" // 清空CLAId，避免引用旧社区的CLA
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
