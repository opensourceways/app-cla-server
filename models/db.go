package models

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

type IDB interface {
	ICLA
}

type ICLA interface {
	CreateCLA(CLA) (CLA, error)
	ListCLA() ([]CLA, error)
}
