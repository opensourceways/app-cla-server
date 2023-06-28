package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	userAdapterInstance            userAdapter
	corpAdminAdatperInstance       corpAdminAdatper
	corpSigningAdapterInstance     corpSigningAdapter
	employeeSigningAdapterInstance employeeSigningAdapter
	employeeManagerAdapterInstance employeeManagerAdapter
	corpEmailDomainAdapterInstance corpEmailDomainAdapter
)

type corpSigningAdapter interface {
	Sign(opt *CorporationSigningCreateOption, linkId string) IModelError
}

type employeeSigningAdapter interface {
	Sign(opt *EmployeeSigning) ([]dbmodels.CorporationManagerListResult, IModelError)
	Remove(csId, esId string) (string, IModelError)
	Update(csId, esId string, enabled bool) (string, IModelError)
	List(csId string) ([]dbmodels.IndividualSigningBasicInfo, IModelError)
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

type corpEmailDomainAdapter interface {
	Add(csId string, opt *CorpEmailDomainCreateOption) IModelError
	List(csId string) ([]string, IModelError)
}

func Init(
	ua userAdapter,
	ca corpAdminAdatper,
	cs corpSigningAdapter,
	es employeeSigningAdapter,
	em employeeManagerAdapter,
	ed corpEmailDomainAdapter,
) {
	userAdapterInstance = ua
	corpAdminAdatperInstance = ca
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
	employeeManagerAdapterInstance = em
	corpEmailDomainAdapterInstance = ed
}
