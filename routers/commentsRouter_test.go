package routers

import (
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// TestRouterRegistration 包内 init 会注册全部路由（含 /v1/individual-signing/cla-confirm/:token），
// 执行到此处即说明无路由冲突 panic；同时校验个人签署相关路由均已注册
// （Router 为相对于 namespace /v1/individual-signing 的模式）。
func TestRouterRegistration(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{"post", "/cla-confirm/:token"},
		{"post", "/:link_id/"},
		{"put", "/:link_id/"},
		{"post", "/:link_id/code"},
		{"get", "/:link_id"},
	}

	for _, c := range cases {
		found := false
		for _, cs := range beego.GlobalControllerRouter["github.com/opensourceways/app-cla-server/controllers:IndividualSigningController"] {
			if cs.Router == c.path && contains(cs.AllowHTTPMethods, c.method) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("route %s %s is not registered", c.method, c.path)
		}
	}
}

func contains(items []string, v string) bool {
	for _, i := range items {
		if i == v {
			return true
		}
	}

	return false
}
