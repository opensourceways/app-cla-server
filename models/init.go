package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	corpAdminAdatperInstance       corpAdminAdatper
	corpSigningAdapterInstance     corpSigningAdapter
	employeeSigningAdapterInstance employeeSigningAdapter
)

type corpSigningAdapter interface {
	Sign(opt *CorporationSigningCreateOption, linkId string) IModelError
}

type employeeSigningAdapter interface {
	Sign(opt *EmployeeSigning) IModelError
}

type corpAdminAdatper interface {
	Add(csId string) (dbmodels.CorporationManagerCreateOption, IModelError)
}

func Init(
	ca corpAdminAdatper,
	cs corpSigningAdapter,
	es employeeSigningAdapter,
) {
	corpAdminAdatperInstance = ca
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
}
