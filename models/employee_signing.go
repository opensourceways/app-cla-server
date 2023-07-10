package models

type EmployeeSigning struct {
	IndividualSigning

	CorpSigningId string `json:"corp_signing_id" required:"true"`
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}
