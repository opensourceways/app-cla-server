// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"fmt"
	"strings"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
	"github.com/beego/beego/v2/core/logs"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/opensourceways/app-cla-server/controllers"
)

func init() {
	setRate()

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

type resp struct {
	Data errResp `json:"data"`
}

type errResp struct {
	ErrCode string `json:"error_code"`
	ErrMsg  string `json:"error_message"`
}

func setRate() {
	requestMaxRate := beego.AppConfig.String("requestLimit")
	requestRate, _ := limiter.NewRateFromFormatted(requestMaxRate)
	l := limiter.New(memory.NewStore(), requestRate)

	beego.InsertFilter("/*", beego.BeforeRouter, func(ctx *context.Context) {
		rateLimit(l, ctx)
	}, true)
}

func rateLimit(limit *limiter.Limiter, ctx *context.Context) {
	if !needCheck(ctx) {
		return
	}

	reached, ip, err := check(limit, ctx)
	if err != nil {
		logs.Error(err)

		return
	}

	if reached {
		logs.Info("too many requests from %s on %s", ip, ctx.Input.URL())

		data := resp{
			Data: errResp{
				ErrCode: "system_error",
				ErrMsg:  "Too Many Requests",
			},
		}

		ctx.Output.JSON(data, false, false)
	}
}

func needCheck(ctx *context.Context) bool {
	url := ctx.Input.URL()
	prefix := []string{
		"/v1/verification-code",
		"/v1/password-retrieval",
	}
	for _, item := range prefix {
		if strings.HasPrefix(url, item) {
			return true
		}
	}

	return false
}

func check(limit *limiter.Limiter, ctx *context.Context) (bool, string, error) {
	opt := limiter.Options{
		IPv4Mask:           limiter.DefaultIPv4Mask,
		IPv6Mask:           limiter.DefaultIPv6Mask,
		TrustForwardHeader: false,
	}

	ip := limiter.GetIP(ctx.Request, opt)

	limiterCtx, err := limit.Get(ctx.Request.Context(), ip.String())
	if err != nil {
		return false, "", fmt.Errorf("fetch limiter ctx failed: %s", err.Error())
	}

	return limiterCtx.Reached, ip.String(), nil
}
