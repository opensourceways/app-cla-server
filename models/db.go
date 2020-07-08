package models

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

func getdb() IDB {
	return db
}

type IDB interface {
	ICLA
}

type ICLA interface {
	CreateCLA(CLA) (CLA, error)
}
