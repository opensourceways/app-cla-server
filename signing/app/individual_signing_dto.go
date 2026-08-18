package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignIndividualCLA struct {
	Link             domain.LinkInfo
	Rep              domain.Representative
	AllSingingInfo   domain.AllSingingInfo
	VerificationCode string
}

func (cmd *CmdToSignIndividualCLA) toIndividualSigning() domain.IndividualSigning {
	return domain.IndividualSigning{
		Link:    cmd.Link,
		Rep:     cmd.Rep,
		Date:    util.Date(),
		AllInfo: cmd.AllSingingInfo,
	}
}

func (cmd *CmdToSignIndividualCLA) toCmd() cmdToCreateCodeForIndividualSigning {
	return cmdToCreateCodeForIndividualSigning{
		Id:        cmd.Link.Id,
		EmailAddr: cmd.Rep.EmailAddr,
	}
}

type CmdToCheckSinging struct {
	LinkId    string
	EmailAddr dp.EmailAddr
}

type CmdToConfirmCLAByToken struct {
	Token string
}

type CLAConfirmResultDTO struct {
	Result      string `json:"result"`
	OrgAlias    string `json:"org_alias"`
	EmailMasked string `json:"email_masked"`

	// audit-only context, never serialized to the API response
	LinkId string `json:"-"`
	ClaId  string `json:"-"`
}

type IndividualSignedDTO struct {
	Type           string      `json:"type"`                      // "individual" 或 "corp"
	Signed         bool        `json:"signed,omitempty"`          // 是否已签署过
	Status         string      `json:"status,omitempty"`          // "valid" 或 "expired"
	VersionMatched bool        `json:"version_matched,omitempty"` // 签署是否当前有效（考虑宽限期）
	DebugInfo      *DebugInfoDTO `json:"_debug,omitempty"`        // 调试信息（仅debug=true时返回）
}

type DebugInfoDTO struct {
	IsLatestClaVersion  bool   `json:"is_latest_cla_version"`     // 你签署的版本是最新的吗？
	InGracePeriod       bool   `json:"in_grace_period"`           // 当前在宽限期内吗？
	GracePeriodDays     int    `json:"grace_period_days"`         // 宽限期是多少天？
	ClaUpdatedAt        string `json:"cla_updated_at"`            // CLA最后更新日期
	DaysSinceLastUpdate int    `json:"days_since_last_update"`    // 自CLA更新以来过了多少天
	GracePeriodEndsAt   string `json:"grace_period_ends_at"`      // 宽限期什么时候结束
}

type CmdToFindSignedCLAInfo = CmdToCheckSinging

type CLAInfoDTO struct {
	CLAId    string `json:"cla_id"`
	Language string `json:"language"`
}
