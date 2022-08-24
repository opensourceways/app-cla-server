// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/ulule/limiter/v3"

	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/ulule/limiter/v3/drivers/store/memory"
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
		beego.NSNamespace("/org-repo",
			beego.NSInclude(
				&controllers.OrgRepoController{},
			),
		),
		beego.NSNamespace("/verification-code",
			beego.NSInclude(
				&controllers.VerificationCodeController{},
			),
		),
	)
	beego.AddNamespace(ns)
}

type rateLimiter struct {
	limiter *limiter.Limiter
}

func runRate() {
	// 限制api的请求次数
	theRateLimit := &rateLimiter{}
	requestMaxRate := beego.AppConfig.String("requestLimit")
	requestRate, _ := limiter.NewRateFromFormatted(requestMaxRate)
	theRateLimit.limiter = limiter.New(memory.NewStore(), requestRate)
	beego.InsertFilter("/*", beego.BeforeRouter, func(ctx *context.Context) {
		rateLimit(theRateLimit, ctx)
	}, true)
}

func rateLimit(rateLimit *rateLimiter, ctx *context.Context) {
	var (
		limiterCtx limiter.Context
		err        error
		req        = ctx.Request
	)
	opt := limiter.Options{
		IPv4Mask:           limiter.DefaultIPv4Mask,
		IPv6Mask:           limiter.DefaultIPv6Mask,
		TrustForwardHeader: false,
	}
	ip := limiter.GetIP(req, opt)
	if strings.HasPrefix(ctx.Input.URL(), "/v1/verification-code") {
		limiterCtx, err = rateLimit.limiter.Get(req.Context(), ip.String())
	}
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	if limiterCtx.Reached {
		logs.Debug("Too Many Requests from %s on %s", ip, ctx.Input.URL())
		ctx.Abort(http.StatusTooManyRequests, "Too Many Requests")
		return
	}
}
