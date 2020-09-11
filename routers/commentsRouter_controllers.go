package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
        beego.ControllerComments{
            Method: "Auth",
            Router: "/:platform/:purpose",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:AuthController"],
        beego.ControllerComments{
            Method: "Get",
            Router: "/authcodeurl/:platform/:purpose",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: "/:uid",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAController"],
        beego.ControllerComments{
            Method: "Get",
            Router: "/:uid",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"],
        beego.ControllerComments{
            Method: "GetSigningPageInfo",
            Router: "/:platform/:org_id/:apply_to",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CLAOrgController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: "/:uid",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
        beego.ControllerComments{
            Method: "Update",
            Router: "/",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationManagerController"],
        beego.ControllerComments{
            Method: "Auth",
            Router: "/auth",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
        beego.ControllerComments{
            Method: "Update",
            Router: "/",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:CorporationSigningController"],
        beego.ControllerComments{
            Method: "SendVerifiCode",
            Router: "/verifi-code",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"],
        beego.ControllerComments{
            Method: "Auth",
            Router: "/auth/:platform",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmailController"],
        beego.ControllerComments{
            Method: "Get",
            Router: "/authcodeurl/:platform",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeManagerController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: "/",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:EmployeeSigningController"],
        beego.ControllerComments{
            Method: "Update",
            Router: "/",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/:cla_org_id",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"],
        beego.ControllerComments{
            Method: "Get",
            Router: "/:cla_org_id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"] = append(beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:OrgSignatureController"],
        beego.ControllerComments{
            Method: "BlankSignature",
            Router: "/blank/:language",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
