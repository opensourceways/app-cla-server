package models

var db IDB

func RegisterDB(idb IDB) {
	db = idb
}

type IDB interface {
	ICLA
}

type ICLA interface {
	CreateCLAMetadata(CLAMetadata) (string, error)
	ListCLAMetadata([]string) ([]CLAMetadata, error)
	GetCLAMetadata(string) (CLAMetadata, error)
	DeleteCLAMetadata(string) error
}
