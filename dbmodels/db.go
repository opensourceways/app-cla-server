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
	IOrgCLA
	IIndividualSigning
	ICLA
	IVerificationCode
	IPDF
}

type ICorporationSigning interface {
	SignAsCorporation(orgCLAID, platform, org, repo string, info CorporationSigningInfo) error
	ListCorporationSigning(CorporationSigningListOption) (map[string][]CorporationSigningDetail, error)
	GetCorporationSigningDetail(platform, org, repo, email string) (string, CorporationSigningDetail, error)
	UploadCorporationSigningPDF(orgCLAID, adminEmail string, pdf []byte) error
	DownloadCorporationSigningPDF(orgCLAID, email string) ([]byte, error)
	CheckCorporationSigning(orgCLAID, email string) (CorporationSigningDetail, error)
	GetCorpSigningInfo(platform, org, repo, email string) (string, *CorporationSigningInfo, error)
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) (map[string]CorporationManagerCheckResult, error)
	AddCorporationManager(orgCLAID string, opt []CorporationManagerCreateOption, managerNumber int) ([]CorporationManagerCreateOption, error)
	DeleteCorporationManager(orgCLAID, role string, emails []string) ([]CorporationManagerCreateOption, error)
	ResetCorporationManagerPassword(string, string, CorporationManagerResetPassword) error
	ListCorporationManager(orgCLAID, email, role string) ([]CorporationManagerListResult, error)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) error
	GetOrgEmailInfo(email string) (OrgEmailCreateInfo, error)
}

type IOrgCLA interface {
	ListOrgCLA(OrgCLAListOption) ([]OrgCLA, error)
	GetOrgCLA(string) (OrgCLA, error)
	CreateOrgCLA(OrgCLA) (string, error)
	DeleteOrgCLA(string) error
}

type IIndividualSigning interface {
	SignAsIndividual(orgCLAID, platform, org, repo string, info IndividualSigningInfo) error
	DeleteIndividualSigning(platform, org, repo, email string) error
	UpdateIndividualSigning(platform, org, repo, email string, enabled bool) error
	IsIndividualSigned(platform, orgID, repoId, email string) (bool, error)
	ListIndividualSigning(opt IndividualSigningListOption) (map[string][]IndividualSigningBasicInfo, error)
}

type ICLA interface {
	CreateCLA(CLA) (string, error)
	ListCLA(CLAListOptions) ([]CLA, error)
	GetCLA(string, bool) (CLA, error)
	DeleteCLA(string) error
	ListCLAByIDs(ids []string) ([]CLA, error)
}

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) error
	GetVerificationCode(opt *VerificationCode) error
}

type IPDF interface {
	UploadOrgSignature(orgCLAID string, pdf []byte) error
	DownloadOrgSignature(orgCLAID string) ([]byte, error)
	DownloadOrgSignatureByMd5(orgCLAID, md5sum string) ([]byte, error)

	UploadBlankSignature(language string, pdf []byte) error
	DownloadBlankSignature(language string) ([]byte, error)
}
