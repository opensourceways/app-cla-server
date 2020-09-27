package dbmodels

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

func GetDB() IDB {
	return db
}

type IDB interface {
	ICorporationSigning
	ICorporationManager
	IOrgEmail
	ICLAOrg
	IIndividualSigning
	ICLA
	IVerifiCode
	IPDF
}

type ICorporationSigning interface {
	SignAsCorporation(string, CorporationSigningInfo) error
	ListCorporationSigning(CorporationSigningListOption) (map[string][]CorporationSigningDetail, error)
	GetCorporationSigningDetail(platform, org, repo, email string) (string, CorporationSigningDetail, error)
	UploadCorporationSigningPDF(claOrgID, adminEmail string, pdf []byte) error
	DownloadCorporationSigningPDF(claOrgID, email string) ([]byte, error)
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) ([]CorporationManagerCheckResult, error)
	AddCorporationManager(claOrgID string, opt []CorporationManagerCreateOption, managerNumber int) error
	DeleteCorporationManager(claOrgID string, opt []CorporationManagerCreateOption) error
	ResetCorporationManagerPassword(string, CorporationManagerResetPassword) error
	ListCorporationManager(claOrgID string, opt CorporationManagerListOption) ([]CorporationManagerListResult, error)
	ListManagersWhenEmployeeSigning(claOrgIDs []string, corporID string) ([]CorporationManagerListResult, error)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) error
	GetOrgEmailInfo(email string) (OrgEmailCreateInfo, error)
}

type ICLAOrg interface {
	ListBindingBetweenCLAAndOrg(CLAOrgListOption) ([]CLAOrg, error)
	ListBindingForSigningPage(CLAOrgListOption) ([]CLAOrg, error)
	GetBindingBetweenCLAAndOrg(string) (CLAOrg, error)
	CreateBindingBetweenCLAAndOrg(CLAOrg) (string, error)
	DeleteBindingBetweenCLAAndOrg(string) error
}

type IIndividualSigning interface {
	SignAsIndividual(string, IndividualSigningInfo) error
	DeleteIndividualSigning(claOrgID, email string) error
	UpdateIndividualSigning(claOrgID, email string, enabled bool) error
	IsIndividualSigned(platform, orgID, repoId, email string) (bool, error)
	ListIndividualSigning(opt IndividualSigningListOption) (map[string][]IndividualSigningBasicInfo, error)
}

type ICLA interface {
	CreateCLA(CLA) (string, error)
	ListCLA(CLAListOptions) ([]CLA, error)
	GetCLA(string) (CLA, error)
	DeleteCLA(string) error
	ListCLAByIDs(ids []string) ([]CLA, error)
}

type IVerifiCode interface {
	CreateVerificationCode(opt VerificationCode) error
	CheckVerificationCode(opt VerificationCode) error
}

type IPDF interface {
	UploadOrgSignature(claOrgID string, pdf []byte) error
	DownloadOrgSignature(claOrgID string) ([]byte, error)

	UploadBlankSignature(language string, pdf []byte) error
	DownloadBlankSignature(language string) ([]byte, error)
}
