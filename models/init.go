package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	userAdapterInstance            userAdapter
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

type userAdapter interface {
	ChangePassword(string, *CorporationManagerResetPassword) IModelError
}

func Init(
	ua userAdapter,
	ca corpAdminAdatper,
	cs corpSigningAdapter,
	es employeeSigningAdapter,
) {
	userAdapterInstance = ua
	corpAdminAdatperInstance = ca
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
}
