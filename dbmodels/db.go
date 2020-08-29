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
	IEmployeeSigning
	IOrgEmail
	ICLAOrg
	IIndividualSigning
	ICLA
}

type ICorporationSigning interface {
	SignAsCorporation(string, CorporationSigningInfo) error
	ListCorporationSigning(CorporationSigningListOption) (map[string][]CorporationSigningDetails, error)
	UpdateCorporationSigning(claOrgID, adminEmail, corporationName string, opt CorporationSigningUpdateInfo) error
}

type ICorporationManager interface {
	CheckCorporationManagerExist(CorporationManagerCheckInfo) (CorporationManagerCheckResult, error)
	AddCorporationManager(claOrgID string, opt []CorporationManagerCreateOption, managerNumber int) error
	DeleteCorporationManager(claOrgID string, opt []CorporationManagerCreateOption) error
	ResetCorporationManagerPassword(string, CorporationManagerResetPassword) error
	ListCorporationManager(claOrgID string, opt CorporationManagerListOption) ([]CorporationManagerListResult, error)
}

type IEmployeeSigning interface {
	SignAsEmployee(claOrgID string, info EmployeeSigningInfo) error
	ListEmployeeSigning(EmployeeSigningListOption) (map[string][]EmployeeSigningInfo, error)
	UpdateEmployeeSigning(claOrgID, email string, opt EmployeeSigningUpdateInfo) error
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
	SignAsIndividual(string, IndividualSigningInfo) error
}

type ICLA interface {
	CreateCLA(CLA) (string, error)
	ListCLA(CLAListOptions) ([]CLA, error)
	GetCLA(string) (CLA, error)
	DeleteCLA(string) error
}
