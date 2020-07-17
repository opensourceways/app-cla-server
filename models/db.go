package models

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

type IDB interface {
	ICLA
	IOrgRepo
}

type ICLA interface {
	CreateCLA(CLA) (string, error)
	ListCLA() ([]CLA, error)
	GetCLA(string) (CLA, error)
}

type IOrgRepo interface {
	CreateOrgRepo(OrgRepo) (string, error)
	DisableOrgRepo(string) error
}
