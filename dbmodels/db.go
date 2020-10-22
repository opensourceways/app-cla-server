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
	IVerificationCode
	IPDF
}

type ICorporationSigning interface {
	SignAsCorporation(claOrgID, platform, org, repo string, info CorporationSigningInfo) error
	ListCorporationSigning(CorporationSigningListOption) (map[string][]CorporationSigningDetail, error)
	GetCorporationSigningDetail(platform, org, repo, email string) (string, CorporationSigningDetail, error)
	UploadCorporationSigningPDF(claOrgID, adminEmail string, pdf []byte) error
	DownloadCorporationSigningPDF(claOrgID, email string) ([]byte, error)
	CheckCorporationSigning(claOrgID, email string) (CorporationSigningDetail, error)
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) (map[string]CorporationManagerCheckResult, error)
	AddCorporationManager(claOrgID string, opt []CorporationManagerCreateOption, managerNumber int) ([]CorporationManagerCreateOption, error)
	DeleteCorporationManager(claOrgID string, opt []CorporationManagerCreateOption) ([]string, error)
	ResetCorporationManagerPassword(string, string, CorporationManagerResetPassword) error
	ListCorporationManager(claOrgID, email, role string) ([]CorporationManagerListResult, error)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) error
	GetOrgEmailInfo(email string) (OrgEmailCreateInfo, error)
}

type ICLAOrg interface {
	ListBindingBetweenCLAAndOrg(CLAOrgListOption) ([]CLAOrg, error)
	GetBindingBetweenCLAAndOrg(string) (CLAOrg, error)
	CreateBindingBetweenCLAAndOrg(CLAOrg) (string, error)
	DeleteBindingBetweenCLAAndOrg(string) error
}

type IIndividualSigning interface {
	SignAsIndividual(claOrgID, platform, org, repo string, info IndividualSigningInfo) error
	DeleteIndividualSigning(platform, org, repo, email string) error
	UpdateIndividualSigning(platform, org, repo, email string, enabled bool) error
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

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) error
	GetVerificationCode(opt *VerificationCode) error
}

type IPDF interface {
	UploadOrgSignature(claOrgID string, pdf []byte) error
	DownloadOrgSignature(claOrgID string) ([]byte, error)

	UploadBlankSignature(language string, pdf []byte) error
	DownloadBlankSignature(language string) ([]byte, error)
}
