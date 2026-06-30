package models

type IndividualSigning struct {
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	CLAId            string          `json:"cla_id"`
	CLALanguage      string          `json:"cla_language"`
	VerificationCode string          `json:"verification_code"`
	Info             TypeSigningInfo `json:"info"`
	PrivacyChecked   bool            `json:"privacy_checked"`
}

type IndividualSigningBasicInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	Enabled bool   `json:"enabled"`
}

type IndividualSigningInfo struct {
	IndividualSigningBasicInfo

	CLAId       string          `json:"cla_id"`
	CLALanguage string          `json:"cla_language"`
	Info        TypeSigningInfo `json:"info"`
}

type IndividualSigned struct {
	Type           string     `json:"type"`                      // "individual" 或 "corp"
	Signed         bool       `json:"signed"`                    // 是否已签署过
	VersionMatched bool       `json:"version_matched,omitempty"` // 签署是否当前有效（考虑宽限期）
	Status         string     `json:"status"`                    // 内部状态："not_signed" | "valid" | "expired"
	DebugInfo      *DebugInfo `json:"_debug,omitempty"`          // 调试信息（仅debug=true时返回）
}

type DebugInfo struct {
	IsLatestClaVersion  bool   `json:"is_latest_cla_version"`     // 你签署的版本是最新的吗？
	InGracePeriod       bool   `json:"in_grace_period"`           // 当前在宽限期内吗？
	GracePeriodDays     int    `json:"grace_period_days"`         // 宽限期是多少天？
	ClaUpdatedAt        string `json:"cla_updated_at"`            // CLA最后更新日期
	DaysSinceLastUpdate int    `json:"days_since_last_update"`    // 自CLA更新以来过了多少天
	GracePeriodEndsAt   string `json:"grace_period_ends_at"`      // 宽限期什么时候结束
}
