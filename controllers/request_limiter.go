package controllers

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

var requestLimiter *requestLimiterImpl

func Init(cfg *Config) error {
	if err := initRequestLimiter(cfg.LimitedAPIs); err != nil {
		return err
	}

	config = *cfg

	return nil
}

func initRequestLimiter(limitedAPIs []string) error {

	requestRate, err := limiter.NewRateFromFormatted(
		fmt.Sprintf("%d-M", config.MaxRequestPerMinute),
	)
	if err != nil {
		return err
	}

	requestLimiter = &requestLimiterImpl{
		limiterImpl: limiter.New(memory.NewStore(), requestRate),
		limitedApis: limitedAPIs,
	}

	adapter.InsertFilter(
		"/*", beego.BeforeRouter,
		requestLimiter.rateLimit,
		true,
	)

	return nil
}

// respData
type respData struct {
	Data interface{} `json:"data"`
}

// errMsg
type errMsg struct {
	ErrCode string `json:"error_code"`
	ErrMsg  string `json:"error_message"`
}

// requestLimiterImpl
type requestLimiterImpl struct {
	limiterImpl *limiter.Limiter
	limitedApis []string
}

func (rl *requestLimiterImpl) rateLimit(ctx *context.Context) {
	if !rl.needCheck(ctx) {
		return
	}

	reached, err := rl.check(ctx)
	if err != nil {
		logs.Error(err)

		return
	}

	if reached {
		data := respData{
			Data: errMsg{
				ErrCode: errTooManyRequest,
				ErrMsg:  "Too Many Requests",
			},
		}

		ctx.Output.JSON(data, false, false)
	}
}

func (rl *requestLimiterImpl) needCheck(ctx *context.Context) bool {
	url := ctx.Input.URL()
	for _, item := range rl.limitedApis {
		if strings.HasPrefix(url, item) {
			return true
		}
	}

	return false
}

func (rl *requestLimiterImpl) check(ctx *context.Context) (bool, error) {
	opt := limiter.Options{
		IPv4Mask:           limiter.DefaultIPv4Mask,
		IPv6Mask:           limiter.DefaultIPv6Mask,
		TrustForwardHeader: false,
	}

	ip := limiter.GetIP(ctx.Request, opt)

	limiterCtx, err := rl.limiterImpl.Get(ctx.Request.Context(), ip.String())
	if err != nil {
		return false, fmt.Errorf("fetch limiter ctx failed: %s", err.Error())
	}

	return limiterCtx.Reached, nil
}
