package platforms

import (
	"context"
	"fmt"

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

func (this *giteeClient) GetAuthorizedEmail() (string, error) {
	es, rs, err := this.c.EmailsApi.GetV5Emails(context.Background(), nil)
	if err != nil {
		if rs.StatusCode == 401 {
			return "", fmt.Errorf(errMsgRefuseToAuthorizeEmail)
		}
		if rs.StatusCode == 403 {
			return "", fmt.Errorf(errMsgNoPublicEmail)
		}
		return "", err
	}

	for _, item := range es {
		if item.State != "confirmed" {
			continue
		}

		for _, scope := range item.Scope {
			if scope == "committed" {
				return item.Email, nil
			}
		}
	}

	return "", fmt.Errorf(errMsgNoPublicEmail)
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

		for _, v := range ls {
			r = append(r, v.Login)
		}

		p += 1
	}

	return r, nil
}
