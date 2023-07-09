package models

import "github.com/opensourceways/app-cla-server/dbmodels"

var (
	claAdapterInstance               claAdapter
	linkAdapterInstance              linkAdapter
	userAdapterInstance              userAdapter
	smtpAdapterInstance              smtpAdapter
	gmailAdapterInstance             gmailAdapter
	corpPDFAdapterInstance           corpPDFAdapter
	corpAdminAdatperInstance         corpAdminAdatper
	accessTokenAdapterInstance       accessTokenAdapter
	corpSigningAdapterInstance       corpSigningAdapter
	employeeSigningAdapterInstance   employeeSigningAdapter
	employeeManagerAdapterInstance   employeeManagerAdapter
	corpEmailDomainAdapterInstance   corpEmailDomainAdapter
	verificationCodeAdapterInstance  verificationCodeAdapter
	individualSigningAdapterInstance individualSigningAdapter
)

type corpSigningAdapter interface {
	Sign(linkId string, opt *CorporationSigningCreateOption) IModelError
	Remove(string) IModelError
	Get(csId string) (CorporationSigning, IModelError)
	List(linkId string) ([]CorporationSigningSummary, IModelError)
	FindCorpSummary(linkId string, email string) (interface{}, IModelError)
}

func RegisterCorpSigningAdapter(a corpSigningAdapter) {
	corpSigningAdapterInstance = a
}

// employeeSigningAdapter
type employeeSigningAdapter interface {
	Sign(opt *EmployeeSigning) ([]dbmodels.CorporationManagerListResult, IModelError)
	Remove(csId, esId string) (string, IModelError)
	Update(csId, esId string, enabled bool) (string, IModelError)
	List(csId string) ([]dbmodels.IndividualSigningBasicInfo, IModelError)
}

func RegisterEmployeeSigningAdapter(a employeeSigningAdapter) {
	employeeSigningAdapterInstance = a
}

// individualSigningAdapter
type individualSigningAdapter interface {
	Sign(linkId string, opt *IndividualSigning) IModelError
	Check(linkId string, email string) (bool, IModelError)
}

func RegisterIndividualSigningAdapter(a individualSigningAdapter) {
	individualSigningAdapterInstance = a
}

// corpAdminAdatper
type corpAdminAdatper interface {
	Add(csId string) (dbmodels.CorporationManagerCreateOption, IModelError)
}

func RegisterCorpAdminAdatper(a corpAdminAdatper) {
	corpAdminAdatperInstance = a
}

// userAdapter
type userAdapter interface {
	ChangePassword(string, *CorporationManagerChangePassword) IModelError
	ResetPassword(linkId string, email string, password string) IModelError
	Logout(userId string)
	Login(opt *CorporationManagerAuthentication) (CorpManagerLoginInfo, IModelError)
	GenKeyForPasswordRetrieval(linkId string, email string) (string, IModelError)
}

func RegisterUserAdapter(a userAdapter) {
	userAdapterInstance = a
}

// employeeManagerAdapter
type employeeManagerAdapter interface {
	Add(string, *EmployeeManagerCreateOption) ([]dbmodels.CorporationManagerCreateOption, IModelError)
	Remove(string, *EmployeeManagerCreateOption) ([]dbmodels.CorporationManagerCreateOption, IModelError)
	List(csId string) ([]dbmodels.CorporationManagerListResult, IModelError)
}

func RegisterEmployeeManagerAdapter(a employeeManagerAdapter) {
	employeeManagerAdapterInstance = a
}

// corpEmailDomainAdapter
type corpEmailDomainAdapter interface {
	Add(csId string, opt *CorpEmailDomainCreateOption) IModelError
	List(csId string) ([]string, IModelError)
}

func RegisterCorpEmailDomainAdapter(a corpEmailDomainAdapter) {
	corpEmailDomainAdapterInstance = a
}

// corpPDFAdapter
type corpPDFAdapter interface {
	Upload(csId string, pdf []byte) IModelError
	Download(csId string) ([]byte, IModelError)
}

func RegisterCorpPDFAdapter(a corpPDFAdapter) {
	corpPDFAdapterInstance = a
}

// verificationCodeAdapter
type verificationCodeAdapter interface {
	CreateForSigning(linkId string, email string) (string, IModelError)
	ValidateForSigning(linkId string, email, code string) IModelError

	CreateForAddingEmailDomain(csId string, email string) (string, IModelError)
	ValidateForAddingEmailDomain(csId string, email, code string) IModelError
}

func RegisterVerificationCodeAdapter(a verificationCodeAdapter) {
	verificationCodeAdapterInstance = a
}

// gmailAdapter
type gmailAdapter interface {
	Authorize(code, scope string) (string, IModelError)
}

func RegisterGmailAdapter(a gmailAdapter) {
	gmailAdapterInstance = a
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
	Add(linkId string, opt *CLACreateOpt, applyTo string) IModelError
	Remove(linkId, claId string) IModelError
	CLALocalFilePath(linkId, claId string) string
	List(linkId string) (dbmodels.CLAOfLink, IModelError)
}

func RegisterCLAAdapter(a claAdapter) {
	claAdapterInstance = a
}

// linkAdapter
type linkAdapter interface {
	Add(submitter string, opt *LinkCreateOption) IModelError
	Remove(linkId string) IModelError
	List(platform string, orgs []string) ([]dbmodels.LinkInfo, IModelError)
	GetLink(linkId string) (org dbmodels.OrgInfo, merr IModelError)
	GetLinkCLA(linkId, claId string) (dbmodels.OrgInfo, dbmodels.CLAInfo, IModelError)
	ListCLAs(linkId, applyTo string) ([]dbmodels.CLADetail, IModelError)
}

func RegisterLinkAdapter(a linkAdapter) {
	linkAdapterInstance = a
}
