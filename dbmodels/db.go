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
	ICLASigning
}

type ICorporationSigning interface {
	InitializeCorpSigning(linkID string, info *OrgInfo, claInfo *CLAInfo) error
	SignAsCorporation(linkID string, info *CorporationSigningOption) *DBError
	IsCorpSigned(linkID, email string) (bool, *DBError)
	ListCorpSignings(linkID, language string) ([]CorporationSigningSummary, *DBError)
	GetCorpSigningBasicInfo(linkID, email string) (*CorporationSigningBasicInfo, *DBError)
	GetCorpSigningDetail(linkID, email string) (*CorporationSigningOption, error)

	UploadCorporationSigningPDF(linkID string, adminEmail string, pdf *[]byte) *DBError
	DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, error)
	IsCorpSigningPDFUploaded(linkID string, email string) (bool, *DBError)
	ListCorpsWithPDFUploaded(linkID string) ([]string, *DBError)
}

type ICorporationManager interface {
	AddCorporationManager(linkID string, opt []CorporationManagerCreateOption, managerNumber int) *DBError
	DeleteCorporationManager(linkID string, emails []string) ([]CorporationManagerCreateOption, *DBError)
	ResetCorporationManagerPassword(linkID, email string, opt CorporationManagerResetPassword) *DBError
	CheckCorporationManagerExist(opt CorporationManagerCheckInfo) (map[string]CorporationManagerCheckResult, error)
	ListCorporationManager(linkID, email, role string) ([]CorporationManagerListResult, *DBError)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) error
	GetOrgEmailInfo(email string) (OrgEmailCreateInfo, *DBError)
}

type IOrgCLA interface {
	HasLink(orgRepo *OrgRepo) (bool, *DBError)
	CreateLink(info *LinkCreateOption) (string, error)
	Unlink(linkID string) error
	ListLinks(opt *LinkListOption) ([]LinkInfo, error)
	GetOrgOfLink(linkID string) (*OrgInfo, *DBError)
}

type IIndividualSigning interface {
	InitializeIndividualSigning(linkID string, info *OrgRepo, claInfo *CLAInfo) error
	SignAsIndividual(linkID string, info *IndividualSigningInfo) *DBError
	DeleteIndividualSigning(linkID, email string) *DBError
	UpdateIndividualSigning(linkID, email string, enabled bool) *DBError
	IsIndividualSigned(orgRepo *OrgRepo, email string) (bool, *DBError)
	ListIndividualSigning(linkID, corpEmail, claLang string) ([]IndividualSigningBasicInfo, *DBError)
}

type ICLASigning interface {
	AddCLAInfo(linkID, applyTo string, info *CLAInfo) error
	GetCLAInfoSigned(linkID, claLang, applyTo string) (*CLAInfo, *DBError)
}

type ICLA interface {
	HasCLA(linkID, applyTo, language string) (bool, error)
	AddCLA(linkID, applyTo string, cla *CLACreateOption) *DBError
	DeleteCLA(linkID, applyTo, language string) *DBError
	GetCLAByType(orgRepo *OrgRepo, applyTo string) (string, []CLADetail, error)
	GetAllCLA(linkID string) (*CLAOfLink, error)
	GetCLAInfoToSign(linkID, claLang, applyTo string) (*CLAInfo, *DBError)
}

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) error
	GetVerificationCode(opt *VerificationCode) *DBError
}

type IPDF interface {
	DownloadOrgSignature(orgCLAID string) ([]byte, error)

	UploadBlankSignature(language string, pdf []byte) error
	DownloadBlankSignature(language string) ([]byte, error)
}
