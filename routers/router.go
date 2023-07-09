// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	beego "github.com/beego/beego/v2/server/web"

	"github.com/opensourceways/app-cla-server/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/cla",
			beego.NSInclude(
				&controllers.CLAController{},
			),
		),
		beego.NSNamespace("/link",
			beego.NSInclude(
				&controllers.LinkController{},
			),
		),
		beego.NSNamespace("/individual-signing",
			beego.NSInclude(
				&controllers.IndividualSigningController{},
			),
		),
		beego.NSNamespace("/employee-signing",
			beego.NSInclude(
				&controllers.EmployeeSigningController{},
			),
		),
		beego.NSNamespace("/employee-manager",
			beego.NSInclude(
				&controllers.EmployeeManagerController{},
			),
		),
		beego.NSNamespace("/corporation-signing",
			beego.NSInclude(
				&controllers.CorporationSigningController{},
			),
		),
		beego.NSNamespace("/corporation-email-domain",
			beego.NSInclude(
				&controllers.CorpEmailDomainController{},
			),
		),
		beego.NSNamespace("/corporation-manager",
			beego.NSInclude(
				&controllers.CorporationManagerController{},
			),
		),
		beego.NSNamespace("/corporation-pdf",
			beego.NSInclude(
				&controllers.CorporationPDFController{},
			),
		),
		beego.NSNamespace("/gmail",
			beego.NSInclude(
				&controllers.GmailController{},
			),
		),
		beego.NSNamespace("/smtp",
			beego.NSInclude(
				&controllers.SMTPController{},
			),
		),
		beego.NSNamespace("/auth",
			beego.NSInclude(
				&controllers.AuthController{},
			),
		),
		beego.NSNamespace("/organization",
			beego.NSInclude(
				&controllers.OrganizationController{},
			),
		),
		beego.NSNamespace("/verification-code",
			beego.NSInclude(
				&controllers.VerificationCodeController{},
			),
		),
		beego.NSNamespace("/password-retrieval",
			beego.NSInclude(
				&controllers.PasswordRetrievalController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
