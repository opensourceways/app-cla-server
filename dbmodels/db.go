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
	IIndividualSigning

	IOrgCLA
	ILink
	ICLA

	IOrgEmail
	IBlankPDF
	IVerificationCode
}

type ICorporationSigning interface {
	SignAsCorporation(orgRepo *OrgRepo, info *CorporationSigningInfo) error
	ListCorporationSigning(orgRepo *OrgRepo, language string) ([]CorporationSigningSummary, error)
	GetCorporationSigningSummary(orgRepo *OrgRepo, email string) (CorporationSigningSummary, error)
	GetCorporationSigningDetail(orgRepo *OrgRepo, email string) (CorporationSigningDetail, error)

	UploadCorporationSigningPDF(orgRepo *OrgRepo, adminEmail string, pdf []byte) error
	DownloadCorporationSigningPDF(orgRepo *OrgRepo, email string) ([]byte, error)
}

type ICorporationManager interface {
	AddCorporationManager(orgRepo *OrgRepo, opt []CorporationManagerCreateOption, managerNumber int) error
	DeleteCorporationManager(orgRepo *OrgRepo, emails []string) ([]CorporationManagerCreateOption, error)
	ResetCorporationManagerPassword(orgRepo *OrgRepo, email string, info *CorporationManagerResetPassword) error
	GetCorpManager(*CorporationManagerCheckInfo) ([]CorporationManagerCheckResult, error)
	ListCorporationManager(orgRepo *OrgRepo, email, role string) ([]CorporationManagerListResult, error)
}

type IIndividualSigning interface {
	SignAsIndividual(orgRepo *OrgRepo, info *IndividualSigningInfo) error
	DeleteIndividualSigning(orgRepo *OrgRepo, email string) error
	UpdateIndividualSigning(orgRepo *OrgRepo, email string, enabled bool) error
	IsIndividualSigned(orgRepo *OrgRepo, email string) (bool, error)
	ListIndividualSigning(orgRepo *OrgRepo, opt *IndividualSigningListOption) ([]IndividualSigningBasicInfo, error)
}

type IOrgEmail interface {
	CreateOrgEmail(opt *OrgEmailCreateInfo) error
	GetOrgEmailInfo(email string) (*OrgEmailCreateInfo, error)
}

type IOrgCLA interface {
	GetOrgCLAWhenSigningAsCorp(orgRepo *OrgRepo, language, signatureMd5 string) (*OrgCLAForSigning, error)
	GetOrgCLAWhenSigningAsIndividual(orgRepo *OrgRepo, language string) (*OrgCLAForSigning, error)
}

type ILink interface {
	CreateLink(info *LinkCreateOption) (string, error)
	Unlink(orgRepo *OrgRepo) error
	ListLinks(opt *LinkListOption) ([]LinkInfo, error)
}

type ICLA interface {
	GetCLAByType(orgRepo *OrgRepo, applyTo string) ([]CLA, error)
	GetAllCLA(orgRepo *OrgRepo) (*CLAOfLink, error)
	AddCLA(orgRepo *OrgRepo, applyTo string, cla *CLA) error
	DeleteCLA(orgRepo *OrgRepo, applyTo, language string) error
	DownloadOrgSignature(orgRepo *OrgRepo, language string) ([]byte, error)
}

type IVerificationCode interface {
	CreateVerificationCode(opt VerificationCode) error
	GetVerificationCode(opt *VerificationCode) error
}

type IBlankPDF interface {
	UploadBlankSignature(language string, pdf []byte) error
	DownloadBlankSignature(language string) ([]byte, error)
}
