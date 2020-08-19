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
}

type ICorporationSigning interface {
	SignAsCorporation(CorporationSigningCreateOption) error
}
