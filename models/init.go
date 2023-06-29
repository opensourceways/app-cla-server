package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	userAdapterInstance             userAdapter
	corpPDFAdapterInstance          corpPDFAdapter
	corpAdminAdatperInstance        corpAdminAdatper
	corpSigningAdapterInstance      corpSigningAdapter
	employeeSigningAdapterInstance  employeeSigningAdapter
	employeeManagerAdapterInstance  employeeManagerAdapter
	corpEmailDomainAdapterInstance  corpEmailDomainAdapter
	verificationCodeAdapterInstance verificationCodeAdapter
)

type corpSigningAdapter interface {
	Sign(opt *CorporationSigningCreateOption, linkId string) IModelError
	Remove(string) IModelError
	Get(csId string) (CorporationSigning, IModelError)
	List(linkId string) ([]CorporationSigningSummary, IModelError)
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

type corpPDFAdapter interface {
	Upload(csId string, pdf []byte) IModelError
	Download(csId string) ([]byte, IModelError)
}

type verificationCodeAdapter interface {
	CreateForSigning(linkId string, email string) (string, IModelError)
	ValidateForSigning(linkId string, email, code string) IModelError

	CreateForAddingEmailDomain(csId string, email string) (string, IModelError)
	ValidateForAddingEmailDomain(csId string, email, code string) IModelError

	CreateForSettingOrgEmail(email string) (string, IModelError)
	ValidateForSettingOrgEmail(email, code string) IModelError

	CreateForPasswordRetrieval(linkId string, email string) (string, IModelError)
	ValidateForPasswordRetrieval(linkId string, email, code string) IModelError
}

func Init(
	ua userAdapter,
	cp corpPDFAdapter,
	ca corpAdminAdatper,
	cs corpSigningAdapter,
	es employeeSigningAdapter,
	em employeeManagerAdapter,
	ed corpEmailDomainAdapter,
	vc verificationCodeAdapter,
) {
	userAdapterInstance = ua
	corpPDFAdapterInstance = cp
	corpAdminAdatperInstance = ca
	corpSigningAdapterInstance = cs
	employeeSigningAdapterInstance = es
	employeeManagerAdapterInstance = em
	corpEmailDomainAdapterInstance = ed
	verificationCodeAdapterInstance = vc
}
