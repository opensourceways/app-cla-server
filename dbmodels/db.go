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
	SignAsCorporation(string, CorporationSigningInfo) error
	ListCorporationsOfOrg(CorporationSigningListOption) (map[string][]CorporationSigningInfo, error)
	UpdateCorporationOfOrg(claOrgID, adminEmail, corporationName string, opt CorporationSigningUpdateInfo) error
}
