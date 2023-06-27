package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	userAdapterInstance            userAdapter
	corpAdminAdatperInstance       corpAdminAdatper
	corpSigningAdapterInstance     corpSigningAdapter
	employeeSigningAdapterInstance employeeSigningAdapter
	employeeManagerAdapterInstance employeeManagerAdapter
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

type employeeManagerAdapter interface {
	Add(string, *EmployeeManagerCreateOption) ([]dbmodels.CorporationManagerCreateOption, IModelError)
	Remove(string, *EmployeeManagerCreateOption) ([]dbmodels.CorporationManagerCreateOption, IModelError)
	List(csId string) ([]dbmodels.CorporationManagerListResult, IModelError)
}

func Init(
	ua userAdapter,
	ca corpAdminAdatper,
	cs corpSigningAdapter,
	es employeeSigningAdapter,
	em employeeManagerAdapter,
) {
	userAdapterInstance = ua
	corpAdminAdatperInstance = ca
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
	employeeManagerAdapterInstance = em
}
