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
	IOrgCLA
	IIndividualSigning
	ICLA
	IVerificationCode
	IPDF
}

type ICorporationSigning interface {
	InitializeCorpSigning(linkID string, info *OrgInfo, cla *CLAInfo) IDBError
	SignCorpCLA(orgCLAID string, info *CorpSigningCreateOpt) IDBError
	IsCorpSigned(linkID, email string) (bool, IDBError)
	ListCorpSignings(linkID, language string) ([]CorporationSigningSummary, IDBError)
	GetCorpSigningDetail(linkID, email string) (*CorpSigningCreateOpt, IDBError)
	GetCorpSigningBasicInfo(linkID, email string) (*CorporationSigningBasicInfo, IDBError)

	UploadCorporationSigningPDF(linkID string, adminEmail string, pdf *[]byte) error
	DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, error)
	IsCorpSigningPDFUploaded(linkID string, email string) (bool, error)
	ListCorpsWithPDFUploaded(linkID string) ([]string, error)
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

type IOrgCLA interface {
	ListOrgs(platform string, orgs []string) ([]OrgCLA, error)
	ListOrgCLA(OrgCLAListOption) ([]OrgCLA, error)
	GetOrgCLA(string) (OrgCLA, error)
	CreateOrgCLA(OrgCLA) (string, error)
	DeleteOrgCLA(string) error
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
	CreateCLA(CLA) (string, error)
	ListCLA(CLAListOptions) ([]CLA, error)
	GetCLA(string, bool) (CLA, error)

	ListCLAByIDs(ids []string) ([]CLA, error)

	GetCLAByType(orgRepo *OrgRepo, applyTo string) (string, []CLADetail, IDBError)
	GetAllCLA(linkID string) (*CLAOfLink, IDBError)
	HasCLA(linkID, applyTo, language string) (bool, IDBError)

	AddCLA(linkID, applyTo string, cla *CLACreateOption) IDBError
	DeleteCLA(linkID, applyTo, language string) IDBError
	DeleteCLAInfo(linkID, applyTo, claLang string) IDBError
	AddCLAInfo(linkID, applyTo string, info *CLAInfo) IDBError
}

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) IDBError
	GetVerificationCode(opt *VerificationCode) IDBError
}

type IPDF interface {
	UploadOrgSignature(orgCLAID string, pdf []byte) error
	DownloadOrgSignature(orgCLAID string) ([]byte, error)
	DownloadOrgSignatureByMd5(orgCLAID, md5sum string) ([]byte, error)
}

type ILink interface {
	GetLinkID(orgRepo *OrgRepo) (string, IDBError)
	CreateLink(info *LinkCreateOption) (string, IDBError)
	Unlink(linkID string) IDBError
	GetOrgOfLink(linkID string) (*OrgInfo, IDBError)
	ListLinks(opt *LinkListOption) ([]LinkInfo, IDBError)
}
