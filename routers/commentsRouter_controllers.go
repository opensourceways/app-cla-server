package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
		beego.ControllerComments{
			Method:           "Auth",
			Router:           "/:platform/:purpose",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           "/authcodeurl/:platform/:purpose",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           "/:link_id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "Add",
			Router:           "/:link_id/:apply_to",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:link_id/:apply_to:/:language",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           "/",
			AllowHTTPMethods: []string{"patch"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Put",
			Router:           "/:link_id/:email",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
		beego.ControllerComments{
			Method:           "Auth",
			Router:           "/auth",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Review",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Preview",
			Router:           "/:link_id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Upload",
			Router:           "/:link_id/:email",
			AllowHTTPMethods: []string{"patch"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationPDFController"],
		beego.ControllerComments{
			Method:           "Download",
			Router:           "/:link_id/:email",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/:link_id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/:link_id/:cla_lang/:cla_hash",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
		beego.ControllerComments{
			Method:           "ResendCorpSigningEmail",
			Router:           "/:link_id/:email",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"],
		beego.ControllerComments{
			Method:           "Auth",
			Router:           "/auth/:platform",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           "/authcodeurl/:platform",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Update",
			Router:           "/:email",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:email",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/:link_id/:cla_lang/:cla_hash",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/:link_id/:cla_lang/:cla_hash",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"],
		beego.ControllerComments{
			Method:           "Check",
			Router:           "/:platform/:org/:repo",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "Link",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "ListLinks",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "Unlink",
			Router:           "/:link_id",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:LinkController"],
		beego.ControllerComments{
			Method:           "GetCLAForSigning",
			Router:           "/:platform/:org_id/:apply_to",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           "/:link_id/:language",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"],
		beego.ControllerComments{
			Method:           "BlankSignature",
			Router:           "/blank/:language",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:VerificationCodeController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/:link_id/:email",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
