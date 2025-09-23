package models

// models/migration.go
type CommunityMigrationOpt struct {
	SourceLinkId string `json:"source_link_id"`
	TargetLinkId string `json:"target_link_id"`
}
