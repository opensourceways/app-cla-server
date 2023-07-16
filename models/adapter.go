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
func AddCLAInstance(linkId string, opt *CLACreateOpt) IModelError {
	return claAdapterInstance.Add(linkId, opt)
}

func CLAFile(linkId, claId string) string {
	return claAdapterInstance.CLALocalFilePath(linkId, claId)
}

func ListCLAInstances(linkId string) (CLAOfLink, IModelError) {
	return claAdapterInstance.List(linkId)
}

func RemoveCLAInstance(linkId, claId string) IModelError {
	return claAdapterInstance.Remove(linkId, claId)
}

// link

func AddLink(submitter string, opt *LinkCreateOption) IModelError {
	return linkAdapterInstance.Add(submitter, opt)
}

func RemoveLink(linkId string) IModelError {
	return linkAdapterInstance.Remove(linkId)
}

func ListLink(platform string, orgs []string) ([]LinkInfo, IModelError) {
	return linkAdapterInstance.List(platform, orgs)
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

func SignCropCLA(linkId string, opt *CorporationSigningCreateOption) IModelError {
	return corpSigningAdapterInstance.Sign(linkId, opt)
}

func RemoveCorpSigning(csId string) IModelError {
	return corpSigningAdapterInstance.Remove(csId)
}

func ListCorpSigning(linkID string) ([]CorporationSigningSummary, IModelError) {
	return corpSigningAdapterInstance.List(linkID)
}

func GetCorpSigning(csId string) (CorporationSigning, IModelError) {
	return corpSigningAdapterInstance.Get(csId)
}

func FindCorpSummary(linkId string, email string) (interface{}, IModelError) {
	return corpSigningAdapterInstance.FindCorpSummary(linkId, email)
}

// corp pdf

func UploadCorpPDF(csId string, pdf []byte) IModelError {
	return corpPDFAdapterInstance.Upload(csId, pdf)
}

func DownloadCorpPDF(csId string) ([]byte, IModelError) {
	return corpPDFAdapterInstance.Download(csId)
}

// employee signing

func VCOfEmployeeSigning(csId, email string) (string, IModelError) {
	return employeeSigningAdapterInstance.Verify(csId, email)
}

func SignEmployeeCLA(opt *EmployeeSigning) ([]CorporationManagerListResult, IModelError) {
	return employeeSigningAdapterInstance.Sign(opt)
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

func SignIndividualCLA(linkId string, opt *IndividualSigning) IModelError {
	return individualSigningAdapterInstance.Sign(linkId, opt)
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
func CreateCorporationAdministratorByAdapter(csId string) (CorporationManagerCreateOption, IModelError) {
	return corpAdminAdatperInstance.Add(csId)
}

// user

func ChangePassword(index string, opt *CorporationManagerChangePassword) IModelError {
	return userAdapterInstance.ChangePassword(index, opt)
}

func CorpManagerLogout(userId string) {
	userAdapterInstance.Logout(userId)
}

func CorpManagerLogin(opt *CorporationManagerLoginInfo) (CorpManagerLoginInfo, IModelError) {
	return userAdapterInstance.Login(opt)
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
