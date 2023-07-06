package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
		beego.ControllerComments{
			Method:           "Callback",
			Router:           `/:platform/:purpose`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
		beego.ControllerComments{
			Method:           "AuthCodeURL",
			Router:           `/authcodeurl/:platform/:purpose`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "Add",
			Router:           `/:link_id/:apply_to`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:link_id/:id`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "DownloadPDF",
			Router:           `/:link_id/:id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorpEmailDomainController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorpEmailDomainController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorpEmailDomainController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorpEmailDomainController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"patch"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Put",
			Router:           `/:link_id/:signing_id`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Auth",
			Router:           `/auth`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Review",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Upload",
			Router:           `/:link_id/:signing_id`,
			AllowHTTPMethods: []string{"patch"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Download",
			Router:           `/:link_id/:signing_id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:link_id/:cla_lang/:cla_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:link_id/:signing_id`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "ResendCorpSigningEmail",
			Router:           `/:link_id/:signing_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "GetCorpInfo",
			Router:           `/:link_id/corps/:email`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "ListDeleted",
			Router:           `/deleted/:link_id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:link_id/:cla_lang/:cla_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Update",
			Router:           `/:signing_id`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:signing_id`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:GmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:GmailController"],
		beego.ControllerComments{
			Method:           "Auth",
			Router:           `/auth`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:GmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:GmailController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/authcodeurl`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"],
		beego.ControllerComments{
			Method:           "Check",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:link_id/:cla_lang/:cla_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "Link",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "ListLinks",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "Unlink",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "GetCLAForSigning",
			Router:           `/:link_id/:apply_to`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "UpdateLinkEmail",
			Router:           `/update/:link_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:PasswordRetrievalController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:PasswordRetrievalController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:PasswordRetrievalController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:PasswordRetrievalController"],
		beego.ControllerComments{
			Method:           "Reset",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"patch"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:TXmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:TXmailController"],
		beego.ControllerComments{
			Method:           "Authorize",
			Router:           `/authorize`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:TXmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:TXmailController"],
		beego.ControllerComments{
			Method:           "Code",
			Router:           `/code`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"],
		beego.ControllerComments{
			Method:           "EmailDomain",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:link_id`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
