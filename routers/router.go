// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/opensourceways/app-cla-server/controllers"
)

func init() {
	runRate()

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
		beego.NSNamespace("/email",
			beego.NSInclude(
				&controllers.EmailController{},
			),
		),
		beego.NSNamespace("/auth",
			beego.NSInclude(
				&controllers.AuthController{},
			),
		),
		beego.NSNamespace("/org-signature",
			beego.NSInclude(
				&controllers.OrgSignatureController{},
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
		beego.NSNamespace("/github",
			beego.NSInclude(
				&controllers.RobotGithubController{},
			),
		),
	)
	beego.AddNamespace(ns)
}

type resp struct {
	Data errResp `json:"data"`
}

type errResp struct {
	ErrCode string `json:"error_code"`
	ErrMsg  string `json:"error_message"`
}

func runRate() {
	requestMaxRate := beego.AppConfig.String("requestLimit")
	requestRate, _ := limiter.NewRateFromFormatted(requestMaxRate)
	l := limiter.New(memory.NewStore(), requestRate)

	beego.InsertFilter("/*", beego.BeforeRouter, func(ctx *context.Context) {
		rateLimit(l, ctx)
	}, true)
}

func rateLimit(l *limiter.Limiter, ctx *context.Context) {
	opt := limiter.Options{
		IPv4Mask:           limiter.DefaultIPv4Mask,
		IPv6Mask:           limiter.DefaultIPv6Mask,
		TrustForwardHeader: false,
	}

	ip := limiter.GetIP(ctx.Request, opt)
	if strings.HasPrefix(ctx.Input.URL(), "/v1/verification-code") {
		limiterCtx, err := l.Get(ctx.Request.Context(), ip.String())
		if err != nil {
			logs.Error("limiter ctx failed: %s", err.Error())
			return
		}

		if limiterCtx.Reached {
			logs.Info("Too Many Requests from %s on %s", ip, ctx.Input.URL())

			data := resp{
				Data: errResp{
					ErrCode: "system_error",
					ErrMsg:  "Too Many Requests",
				},
			}

			ctx.Output.JSON(data, false, false)

			return
		}
	}

}
