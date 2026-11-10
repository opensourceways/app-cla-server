package models

import (
	"bytes"
	"encoding/csv"

	"github.com/opensourceways/app-cla-server/util"
)

type TypeSigningInfo map[string]string

type CorporationSigning struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}

type CorporationSigningCreateOption struct {
	CLAId            string          `json:"cla_id"`
	CLALanguage      string          `json:"cla_language"`
	AdminName        string          `json:"admin_name"`
	AdminEmail       string          `json:"admin_email"`
	CorporationName  string          `json:"corporation_name"`
	VerificationCode string          `json:"verification_code"`
	Info             TypeSigningInfo `json:"info"`
	PrivacyChecked   bool            `json:"privacy_checked"`
}

func (opt *CorporationSigningCreateOption) ToCorporationSigning() CorporationSigning {
	return CorporationSigning{
		CorporationSigningBasicInfo: CorporationSigningBasicInfo{
			CLAId:           opt.CLAId,
			CLALanguage:     opt.CLALanguage,
			AdminName:       opt.AdminName,
			AdminEmail:      opt.AdminEmail,
			CorporationName: opt.CorporationName,
			Date:            util.Date(),
		},
		Info: opt.Info,
	}
}

type CorporationSigningPageSummary struct {
	Total int64                       `json:"total"`
	Data  []CorporationSigningSummary `json:"page_data"`
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	Id             string `json:"id"`
	AdminAdded     bool   `json:"admin_added"`
	PDFUploaded    bool   `json:"pdf_uploaded"`
	AdminAddedDate string `json:"admin_added_date"`
}

type CorporationSigningBasicInfo struct {
	CLAId           string `json:"cla_id"`
	CLALanguage     string `json:"cla_language"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

// models/corp_signing.go
type RepresentativeUpdateOption struct {
	RepName  string `json:"rep_name" valid:"Required"`
	RepEmail string `json:"rep_email" valid:"Required;Email"`
}

func UpdateCorpRepresentative(userId, linkID, signingID string, opt *RepresentativeUpdateOption) IModelError {
	return corpSigningAdapterInstance.UpdateRepresentative(userId, linkID, signingID, opt)
}

type CorporationSigningPendingItem struct {
	SigningId         string `json:"signing_id"`
	CorpName          string `json:"corp_name"`
	AdminEmail        string `json:"admin_email"`
	SignedCLAVersion  string `json:"signed_cla_version"`
	PendingCLAVersion string `json:"pending_cla_version"`
	NotifyCount       int    `json:"notify_count"`
	LastNotifyTime    int64  `json:"last_notify_time"`
}

func FindPendingAgreements(userId, linkId string) ([]CorporationSigningPendingItem, IModelError) {
	return corpSigningAdapterInstance.FindPendingAgreements(userId, linkId)
}

// CorpAutoApprovalOption is the request/response body for the auto-approval
// preference of a corporation signing.
type CorpAutoApprovalOption struct {
	Enabled bool `json:"enabled"`
}

// ExportCorpSigningCSV exports the corp signing list as CSV bytes.
// adminAdded filters: true = only completed, false = only not-completed.
func ExportCorpSigningCSV(userId, linkID string, adminAdded bool) ([]byte, IModelError) {
	list, err := corpSigningAdapterInstance.List(userId, linkID)
	if err != nil {
		return nil, err
	}

	filtered := make([]CorporationSigningSummary, 0, len(list))
	for i := range list {
		if list[i].AdminAdded == adminAdded {
			filtered = append(filtered, list[i])
		}
	}

	return generateCorpSigningCSV(filtered), nil
}

func generateCorpSigningCSV(items []CorporationSigningSummary) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	w.Write([]string{"企业名称", "管理员邮箱", "管理员姓名", "申请日期", "状态", "审批完成日期", "CLA语言"})

	status := func(b bool) string {
		if b {
			return "已完成"
		}
		return "未完成"
	}

	for i := range items {
		item := &items[i]
		w.Write([]string{
			item.CorporationName,
			item.AdminEmail,
			item.AdminName,
			item.Date,
			status(item.AdminAdded),
			item.AdminAddedDate,
			item.CLALanguage,
		})
	}

	w.Flush()
	return buf.Bytes()
}
