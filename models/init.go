package models

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

var (
	claAdapterInstance               claAdapter
	linkAdapterInstance              linkAdapter
	userAdapterInstance              userAdapter
	smtpAdapterInstance              smtpAdapter
	corpPDFAdapterInstance           corpPDFAdapter
	corpAdminAdatperInstance         corpAdminAdatper
	accessTokenAdapterInstance       accessTokenAdapter
	corpSigningAdapterInstance       corpSigningAdapter
	employeeSigningAdapterInstance   employeeSigningAdapter
	employeeManagerAdapterInstance   employeeManagerAdapter
	corpEmailDomainAdapterInstance   corpEmailDomainAdapter
	individualSigningAdapterInstance individualSigningAdapter
)

type corpSigningAdapter interface {
	Verify(linkId, email string) (string, IModelError)
	Sign(linkId string, opt *CorporationSigningCreateOption, claFields []CLAField) IModelError
	Remove(string, string) IModelError
	Get(userId, csId string, email dp.EmailAddr) (string, CorporationSigning, IModelError)
	List(userId, linkId string) ([]CorporationSigningSummary, IModelError)
	FindCorpSummary(linkId string, email string) (interface{}, IModelError)
}

func RegisterCorpSigningAdapter(a corpSigningAdapter) {
	corpSigningAdapterInstance = a
}

// employeeSigningAdapter
type employeeSigningAdapter interface {
	Verify(csId, email string) (string, IModelError)
	Sign(opt *EmployeeSigning, claFields []CLAField) ([]CorporationManagerListResult, IModelError)
	Remove(csId, esId string) (string, IModelError)
	Update(csId, esId string, enabled bool) (string, IModelError)
	List(csId string) ([]IndividualSigningBasicInfo, IModelError)
}

func RegisterEmployeeSigningAdapter(a employeeSigningAdapter) {
	employeeSigningAdapterInstance = a
}

// individualSigningAdapter
type individualSigningAdapter interface {
	Verify(linkId, email string) (string, IModelError)
	Sign(linkId string, opt *IndividualSigning, claFields []CLAField) IModelError
	Check(linkId string, email string) (bool, IModelError)
}

func RegisterIndividualSigningAdapter(a individualSigningAdapter) {
	individualSigningAdapterInstance = a
}

// corpAdminAdatper
type corpAdminAdatper interface {
	Add(userId, csId string) (string, CorporationManagerCreateOption, IModelError)
}

func RegisterCorpAdminAdatper(a corpAdminAdatper) {
	corpAdminAdatperInstance = a
}

// userAdapter
type userAdapter interface {
	Login(opt *CorporationManagerLoginInfo) (CorpManagerLoginInfo, IModelError)
	GetUserInfo(string) (CorpManagerUserInfo, IModelError)
	ChangePassword(string, *CorporationManagerChangePassword) IModelError
	ResetPassword(linkId string, email string, password []byte) IModelError
	GenKeyForPasswordRetrieval(linkId string, email string) (string, IModelError)
}

func RegisterUserAdapter(a userAdapter) {
	userAdapterInstance = a
}

// employeeManagerAdapter
type employeeManagerAdapter interface {
	Add(string, *EmployeeManagerCreateOption) ([]CorporationManagerCreateOption, IModelError)
	Remove(string, *EmployeeManagerDeleteOption) ([]CorporationManagerCreateOption, IModelError)
	List(csId string) ([]CorporationManagerListResult, IModelError)
}

func RegisterEmployeeManagerAdapter(a employeeManagerAdapter) {
	employeeManagerAdapterInstance = a
}

// corpEmailDomainAdapter
type corpEmailDomainAdapter interface {
	Verify(csId string, email string) (string, IModelError)
	Add(csId string, opt *CorpEmailDomainCreateOption) IModelError
	List(csId string) ([]string, IModelError)
}

func RegisterCorpEmailDomainAdapter(a corpEmailDomainAdapter) {
	corpEmailDomainAdapterInstance = a
}

// corpPDFAdapter
type corpPDFAdapter interface {
	Upload(userId, csId string, pdf []byte) IModelError
	Download(userId, csId string, email dp.EmailAddr) ([]byte, IModelError)
}

func RegisterCorpPDFAdapter(a corpPDFAdapter) {
	corpPDFAdapterInstance = a
}

// smtpAdapter
type smtpAdapter interface {
	Verify(opt *EmailAuthorizationReq) (string, IModelError)
	Authorize(opt *EmailAuthorization) IModelError
}

func RegisterSMTPAdapter(a smtpAdapter) {
	smtpAdapterInstance = a
}

// accessTokenAdapter
type accessTokenAdapter interface {
	Remove(string)
	Add(payload []byte) (AccessToken, IModelError)
	ValidateAndRefresh(AccessToken) (AccessToken, []byte, IModelError)
}

func RegisterAccessTokenAdapter(at accessTokenAdapter) {
	accessTokenAdapterInstance = at
}

// claAdapter
type claAdapter interface {
	Add(userId, linkId string, opt *CLACreateOpt) IModelError
	Remove(userId, linkId, claId string) IModelError
	CLALocalFilePath(linkId, claId string) string
	List(userId, linkId string) (CLAOfLink, IModelError)
}

func RegisterCLAAdapter(a claAdapter) {
	claAdapterInstance = a
}

// linkAdapter
type linkAdapter interface {
	Add(submitter string, opt *LinkCreateOption) IModelError
	Remove(userId, linkId string) IModelError
	List(userId string) ([]LinkInfo, IModelError)
	GetLink(linkId string) (org OrgInfo, merr IModelError)
	GetLinkCLA(linkId, claId string) (OrgInfo, CLAInfo, IModelError)
	ListCLAs(linkId, applyTo string) ([]CLADetail, IModelError)
}

func RegisterLinkAdapter(a linkAdapter) {
	linkAdapterInstance = a
}
