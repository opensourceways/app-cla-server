package platforms

import (
	"net/http"

	"github.com/opensourceways/app-cla-server/util"
)

type loginInfo struct {
	Login string `json:"login"`
}

func getUser(req *http.Request) (string, error) {
	var v loginInfo
	c := util.NewHttpClient(3)

	if _, err := c.ForwardTo(req, &v); err != nil {
		return "", err
	}

	return v.Login, nil
}

func listOrg(req *http.Request) ([]string, error) {
	var v []loginInfo
	c := util.NewHttpClient(3)

	if _, err := c.ForwardTo(req, &v); err != nil {
		return nil, err
	}

	r := make([]string, len(v))
	for i := range v {
		r[i] = v[i].Login
	}

	return r, nil
}
