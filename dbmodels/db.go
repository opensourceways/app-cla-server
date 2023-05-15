package dbmodels

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

func GetDB() IDB {
	return db
}

type IDB interface {
	ILink
	ICorporationSigning
	ICorporationManager
	IOrgEmail
	IIndividualSigning
	ICLA
	IVerificationCode
	Close() error
}

type ICorporationSigning interface {
	InitializeCorpSigning(linkID string, info *OrgInfo, cla *CLAInfo) IDBError
	SignCorpCLA(orgCLAID string, info *CorpSigningCreateOpt) IDBError
	DeleteCorpSigning(*SigningIndex) IDBError
	IsCorpSigned(string, string) (bool, IDBError)
	ListCorpSignings(linkID, language string) ([]CorporationSigningSummary, IDBError)
	ListDeletedCorpSignings(linkID string) ([]CorporationSigningBasicInfo, IDBError)
	GetCorpSigningDetail(*SigningIndex) (*CLAInfo, *CorpSigningCreateOpt, IDBError)
	GetCorpSigningBasicInfo(*SigningIndex) (*CorporationSigningBasicInfo, IDBError)
	AddCorpEmailDomain(index *SigningIndex, domain string) IDBError
	GetCorpEmailDomains(*SigningIndex) ([]string, IDBError)

	UploadCorporationSigningPDF(*SigningIndex, *[]byte) IDBError
	DownloadCorporationSigningPDF(*SigningIndex) (*[]byte, IDBError)
	IsCorpSigningPDFUploaded(*SigningIndex) (bool, IDBError)
	ListCorpsWithPDFUploaded(linkID string) ([]string, IDBError)
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) (map[string]CorporationManagerCheckResult, IDBError)
	AddCorpAdministrator(*SigningIndex, *CorporationManagerCreateOption) IDBError
	AddEmployeeManager(*SigningIndex, *CorporationManagerCreateOption) IDBError
	DeleteEmployeeManager(orgCLAID string, emails []string) ([]CorporationManagerCreateOption, IDBError)
	ResetCorporationManagerPassword(string, string, CorporationManagerResetPassword) IDBError
	GetCorporationDetail(index *SigningIndex) (CorporationDetail, IDBError)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) IDBError
	GetOrgEmailInfo(email string) (*OrgEmailCreateInfo, IDBError)
}

type IIndividualSigning interface {
	InitializeIndividualSigning(linkID string, info *CLAInfo) IDBError
	SignIndividualCLA(linkID string, info *IndividualSigningInfo) IDBError
	DeleteIndividualSigning(index *SigningIndex) (IndividualSigningBasicInfo, IDBError)
	IsIndividualSigned(linkID, email string) (bool, IDBError)
	ListIndividualSigning(linkID, claLang string) ([]IndividualSigningBasicInfo, IDBError)

	SignEmployeeCLA(*SigningIndex, *IndividualSigningInfo) IDBError
	ListEmployeeSigning(index *SigningIndex, claLang string) ([]IndividualSigningBasicInfo, IDBError)
	UpdateEmployeeSigning(index *SigningIndex, enabled bool) (IndividualSigningBasicInfo, IDBError)

	GetCLAInfoSigned(linkID, claLang, applyTo string) (*CLAInfo, IDBError)
}

type ICLA interface {
	GetCLAByType(linkID, applyTo string) ([]CLADetail, IDBError)
	GetAllCLA(linkID string) (*CLAOfLink, IDBError)
	HasCLA(linkID, applyTo, language string) (bool, IDBError)
	DownloadCorpCLAPDF(linkID, lang string) ([]byte, IDBError)

	AddCLA(linkID, applyTo string, cla *CLACreateOption) IDBError
	DeleteCLA(linkID, applyTo, language string) IDBError
	DeleteCLAInfo(linkID, applyTo, claLang string) IDBError
	AddCLAInfo(linkID, applyTo string, info *CLAInfo) IDBError
	GetCLAInfoToSign(linkID, claLang, applyTo string) (*CLAInfo, IDBError)

	UploadCLAPDF(key CLAPDFIndex, pdf []byte) IDBError
	DownloadCLAPDF(key CLAPDFIndex) ([]byte, IDBError)
	DeleteCLAPDF(key CLAPDFIndex) IDBError
}

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) IDBError
	GetVerificationCode(opt *VerificationCode) IDBError
}

type ILink interface {
	GetLinkID(orgRepo *OrgRepo) (string, IDBError)
	CreateLink(info *LinkCreateOption) (string, IDBError)
	UpdateLinkEmail(info *LinkCreateOption) IDBError
	Unlink(linkID string) IDBError
	GetOrgOfLink(linkID string) (*OrgInfo, IDBError)
	ListLinks(opt *LinkListOption) ([]LinkInfo, IDBError)
	GetAllLinks() ([]LinkInfo, IDBError)
}
