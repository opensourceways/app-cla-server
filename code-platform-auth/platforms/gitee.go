package platforms

import (
	"context"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/antihax/optional"
	"golang.org/x/oauth2"
)

type giteeClient struct {
	accessToken  string
	refreshToken string
	c            *gitee.APIClient
}

func newGiteeClient(accessToken, refreshToken string) *giteeClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})

	conf := gitee.NewConfiguration()
	conf.HTTPClient = oauth2.NewClient(context.Background(), ts)

	cli := gitee.NewAPIClient(conf)

	return &giteeClient{refreshToken: refreshToken, accessToken: accessToken, c: cli}
}

func (this *giteeClient) GetUser() (string, error) {
	u, _, err := this.c.UsersApi.GetV5User(context.Background(), nil)
	if err != nil {
		return "", err
	}
	return u.Login, err
}

func (this *giteeClient) IsOrgExist(org string) (bool, error) {
	orgs, err := this.ListOrg()
	if err != nil {
		//TODO :is token expiry
		return false, err
	}

	for _, item := range orgs {
		if item == org {
			return true, nil
		}
	}
	return false, nil
}

func (this *giteeClient) ListOrg() ([]string, error) {
	var r []string

	p := int32(1)
	opt := gitee.GetV5UserOrgsOpts{Admin: optional.NewBool(true)}
	for {
		opt.Page = optional.NewInt32(p)
		ls, _, err := this.c.OrganizationsApi.GetV5UserOrgs(context.Background(), &opt)
		if err != nil {
			return nil, err
		}

		if len(ls) == 0 {
			break
		}

		p += 1

		for _, v := range ls {
			r = append(r, v.Login)
		}
	}

	return r, nil
}
