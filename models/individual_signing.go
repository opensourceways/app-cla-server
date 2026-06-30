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
	Signed         bool       `json:"signed,omitempty"`          // 是否已签署过
	VersionMatched bool       `json:"version_matched,omitempty"` // 签署是否当前有效（考虑宽限期）
	Status         string     `json:"status,omitempty"`          // 内部状态："not_signed" | "valid" | "expired"
	DebugInfo      *DebugInfo `json:"_debug,omitempty"`          // 调试信息（仅debug=true时返回）
}

type DebugInfo struct {
	IsLatestClaVersion bool   `json:"is_latest_cla_version"`  // 签署的CLA版本是否最新？
	InGracePeriod      bool   `json:"in_grace_period"`        // 是否在宽限期内？
	GracePeriodDays    int    `json:"grace_period_days"`      // 宽限期总共多少天
	ClaUpdatedAt       string `json:"cla_updated_at"`         // CLA最后更新的日期
	ElapsedDays        int    `json:"elapsed_days"`           // 自CLA更新以来已经过了多少天
	GracePeriodEndsAt  string `json:"grace_period_ends_at"`   // 宽限期结束日期
}
