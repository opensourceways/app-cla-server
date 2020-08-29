package models

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

type IDB interface {
	ICLA
	IOrgRepo
	IIndividualSigning
}

type ICLA interface {
	CreateCLA(CLA) (string, error)
	ListCLA(CLAListOptions) ([]CLA, error)
	GetCLA(string) (CLA, error)
	DeleteCLA(string) error

	CreateCLAMetadata(CLAMetadata) (string, error)
	ListCLAMetadata([]string) ([]CLAMetadata, error)
	GetCLAMetadata(string) (CLAMetadata, error)
	DeleteCLAMetadata(string) error
}

type IOrgRepo interface {
	CreateOrgRepo(OrgRepo) (string, error)
	DisableOrgRepo(string) error
	ListOrgRepo(OrgRepos) ([]OrgRepo, error)
}

type IIndividualSigning interface {
	SignAsIndividual(IndividualSigning) error
}

type ICorporationSigning interface {
	SignAsCorporation(CorporationSigning) error
}
