package models

type AccessToken struct {
	Id   string
	CSRF string
}

func NewAccessToken(payload []byte) (AccessToken, IModelError) {
	return accessTokenAdapterInstance.Add(payload)
}

func RemoveAccessToken(tokenId string) {
	accessTokenAdapterInstance.Remove(tokenId)
}

func ValidateAndRefreshAccessToken(token AccessToken) (AccessToken, []byte, IModelError) {
	return accessTokenAdapterInstance.ValidateAndRefresh(token)
}

// cla
func AddCLAInstance(userId, linkId string, opt *CLACreateOpt) IModelError {
	return claAdapterInstance.Add(userId, linkId, opt)
}

func CLAFile(linkId, claId string) string {
	return claAdapterInstance.CLALocalFilePath(linkId, claId)
}

func ListCLAInstances(userId, linkId string) (CLAOfLink, IModelError) {
	return claAdapterInstance.List(userId, linkId)
}

func RemoveCLAInstance(userId, linkId, claId string) IModelError {
	return claAdapterInstance.Remove(userId, linkId, claId)
}

// link

func AddLink(submitter string, opt *LinkCreateOption) IModelError {
	return linkAdapterInstance.Add(submitter, opt)
}

func RemoveLink(userId, linkId string) IModelError {
	return linkAdapterInstance.Remove(userId, linkId)
}

func ListLink(userId string) ([]LinkInfo, IModelError) {
	return linkAdapterInstance.List(userId)
}

func GetLinkCLA(linkId, claId string) (OrgInfo, CLAInfo, IModelError) {
	return linkAdapterInstance.GetLinkCLA(linkId, claId)
}

func ListCLAs(linkId, applyTo string) ([]CLADetail, IModelError) {
	return linkAdapterInstance.ListCLAs(linkId, applyTo)
}

func GetLink(linkId string) (OrgInfo, IModelError) {
	return linkAdapterInstance.GetLink(linkId)
}

// corp signing

func VCOfCorpSigning(linkId, email string) (string, IModelError) {
	return corpSigningAdapterInstance.Verify(linkId, email)
}

func SignCropCLA(linkId string, opt *CorporationSigningCreateOption, claFields []CLAField) IModelError {
	return corpSigningAdapterInstance.Sign(linkId, opt, claFields)
}

func RemoveCorpSigning(userId, csId string) IModelError {
	return corpSigningAdapterInstance.Remove(userId, csId)
}

func ListCorpSigning(userId, linkID string) ([]CorporationSigningSummary, IModelError) {
	return corpSigningAdapterInstance.List(userId, linkID)
}

func GetCorpSigning(userId, csId string) (string, CorporationSigning, IModelError) {
	return corpSigningAdapterInstance.Get(userId, csId)
}

func FindCorpSummary(linkId string, email string) (interface{}, IModelError) {
	return corpSigningAdapterInstance.FindCorpSummary(linkId, email)
}

// corp pdf

func UploadCorpPDF(userId, csId string, pdf []byte) IModelError {
	return corpPDFAdapterInstance.Upload(userId, csId, pdf)
}

func DownloadCorpPDF(userId, csId string) ([]byte, IModelError) {
	return corpPDFAdapterInstance.Download(userId, csId)
}

// employee signing

func VCOfEmployeeSigning(csId, email string) (string, IModelError) {
	return employeeSigningAdapterInstance.Verify(csId, email)
}

func SignEmployeeCLA(opt *EmployeeSigning, claFields []CLAField) ([]CorporationManagerListResult, IModelError) {
	return employeeSigningAdapterInstance.Sign(opt, claFields)
}

func UpdateEmployeeSigning(csId, esId string, enabled bool) (string, IModelError) {
	return employeeSigningAdapterInstance.Update(csId, esId, enabled)
}

func ListEmployeeSignings(csId string) ([]IndividualSigningBasicInfo, IModelError) {
	return employeeSigningAdapterInstance.List(csId)
}

func RemoveEmployeeSigning(csId, esId string) (string, IModelError) {
	return employeeSigningAdapterInstance.Remove(csId, esId)
}

// employee manager

func ListEmployeeManagers(csId string) ([]CorporationManagerListResult, IModelError) {
	return employeeManagerAdapterInstance.List(csId)
}

func AddEmployeeManager(csId string, opt *EmployeeManagerCreateOption) (
	[]CorporationManagerCreateOption, IModelError,
) {
	return employeeManagerAdapterInstance.Add(csId, opt)
}

func RemoveEmployeeManager(csId string, opt *EmployeeManagerDeleteOption) (
	[]CorporationManagerCreateOption, IModelError,
) {
	return employeeManagerAdapterInstance.Remove(csId, opt)
}

// individual signing

func VCOfIndividualSigning(linkId, email string) (string, IModelError) {
	return individualSigningAdapterInstance.Verify(linkId, email)
}

func SignIndividualCLA(linkId string, opt *IndividualSigning, claFields []CLAField) IModelError {
	return individualSigningAdapterInstance.Sign(linkId, opt, claFields)
}

func CheckSigning(linkId string, email string) (bool, IModelError) {
	return individualSigningAdapterInstance.Check(linkId, email)
}

// email domain
func VerifyCorpEmailDomain(csId string, email string) (string, IModelError) {
	return corpEmailDomainAdapterInstance.Verify(csId, email)
}

func AddCorpEmailDomain(csId string, opt *CorpEmailDomainCreateOption) IModelError {
	return corpEmailDomainAdapterInstance.Add(csId, opt)
}

func ListCorpEmailDomains(csId string) ([]string, IModelError) {
	return corpEmailDomainAdapterInstance.List(csId)
}

// corp admin
func CreateCorporationAdministratorByAdapter(userId, csId string) (string, CorporationManagerCreateOption, IModelError) {
	return corpAdminAdatperInstance.Add(userId, csId)
}

// user

func ChangePassword(index string, opt *CorporationManagerChangePassword) IModelError {
	return userAdapterInstance.ChangePassword(index, opt)
}

func CorpManagerLogin(opt *CorporationManagerLoginInfo) (CorpManagerLoginInfo, IModelError) {
	return userAdapterInstance.Login(opt)
}

func GetUserInfo(userId string) (CorpManagerUserInfo, IModelError) {
	return userAdapterInstance.GetUserInfo(userId)
}

// org email

func VerifySMTPEmail(opt *EmailAuthorizationReq) (string, IModelError) {
	return smtpAdapterInstance.Verify(opt)
}

func AuthorizeSMTPEmail(opt *EmailAuthorization) IModelError {
	return smtpAdapterInstance.Authorize(opt)
}

// password retrivieal

func GenKeyForPasswordRetrieval(linkId string, opt *PasswordRetrievalKey) (string, IModelError) {
	return userAdapterInstance.GenKeyForPasswordRetrieval(
		linkId, opt.Email,
	)
}

func ResetPassword(linkId string, opt *PasswordRetrieval, key string) IModelError {
	return userAdapterInstance.ResetPassword(linkId, key, opt.Password)
}
