package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningBasicInfo struct {
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	PDFUploaded bool `json:"pdf_uploaded"`
}

type CorporationSigningDetail struct {
	CorporationSigningSummary

	Info TypeSigningInfo `json:"info"`
}

type CorporationSigningInfo struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}
