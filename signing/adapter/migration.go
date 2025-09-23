package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
)

func NewMigrationAdapter(s app.MigrationService) *migrationAdapter {
	return &migrationAdapter{
		s: s,
	}
}

type migrationAdapter struct {
	s app.MigrationService
}

func (adapter *migrationAdapter) MigrateCommunityData(userId string, opt *models.CommunityMigrationOpt) models.IModelError {
	cmd := app.CmdToMigrateCommunity{
		UserId:       userId,
		SourceLinkId: opt.SourceLinkId,
		TargetLinkId: opt.TargetLinkId,
	}

	if err := adapter.s.MigrateCommunityData(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}
