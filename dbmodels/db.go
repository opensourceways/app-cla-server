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
}

type ICorporationSigning interface {
	InitializeCorpSigning(linkID string, info *OrgInfo, cla *CLAInfo) IDBError
	SignCorpCLA(orgCLAID string, info *CorpSigningCreateOpt) IDBError
	DeleteCorpSigning(linkID, email string) IDBError
	IsCorpSigned(linkID, email string) (bool, IDBError)
	ListCorpSignings(linkID, language string) ([]CorporationSigningSummary, IDBError)
	ListDeletedCorpSignings(linkID string) ([]CorporationSigningBasicInfo, IDBError)
	GetCorpSigningDetail(linkID, email string) (*CLAInfo, *CorpSigningCreateOpt, IDBError)
	GetCorpSigningBasicInfo(linkID, email string) (*CorporationSigningBasicInfo, IDBError)

	UploadCorporationSigningPDF(linkID string, adminEmail string, pdf *[]byte) IDBError
	DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, IDBError)
	IsCorpSigningPDFUploaded(linkID string, email string) (bool, IDBError)
	ListCorpsWithPDFUploaded(linkID string) ([]string, IDBError)
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) (map[string]CorporationManagerCheckResult, IDBError)
	AddCorpAdministrator(linkID string, opt *CorporationManagerCreateOption) IDBError
	AddEmployeeManager(linkID string, opt []CorporationManagerCreateOption) IDBError
	DeleteEmployeeManager(orgCLAID string, emails []string) ([]CorporationManagerCreateOption, IDBError)
	ResetCorporationManagerPassword(string, string, CorporationManagerResetPassword) IDBError
	ListCorporationManager(orgCLAID, email, role string) ([]CorporationManagerListResult, IDBError)
}

type IOrgEmail interface {
	CreateOrgEmail(opt OrgEmailCreateInfo) IDBError
	GetOrgEmailInfo(email string) (*OrgEmailCreateInfo, IDBError)
}

type IIndividualSigning interface {
	InitializeIndividualSigning(linkID string, info *CLAInfo) IDBError
	SignIndividualCLA(linkID string, info *IndividualSigningInfo) IDBError
	DeleteIndividualSigning(linkID, email string) IDBError
	UpdateIndividualSigning(linkID, email string, enabled bool) IDBError
	IsIndividualSigned(linkID, email string) (bool, IDBError)
	ListIndividualSigning(linkID, corpEmail, claLang string) ([]IndividualSigningBasicInfo, IDBError)

	GetCLAInfoSigned(linkID, claLang, applyTo string) (*CLAInfo, IDBError)
}

type ICLA interface {
	GetCLAByType(orgRepo *OrgRepo, applyTo string) (string, []CLADetail, IDBError)
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
	Unlink(linkID string) IDBError
	GetOrgOfLink(linkID string) (*OrgInfo, IDBError)
	ListLinks(opt *LinkListOption) ([]LinkInfo, IDBError)
	GetAllLinks() ([]LinkInfo, IDBError)
}
