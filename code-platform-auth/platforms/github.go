package platforms

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type githubClient struct {
	accessToken  string
	refreshToken string
	c            *github.Client
}

func newGithubClient(accessToken, refreshToken string) *githubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(context.Background(), ts)
	cli := github.NewClient(tc)

	return &githubClient{refreshToken: refreshToken, accessToken: accessToken, c: cli}
}

func (this *githubClient) GetUser() (string, error) {
	u, _, err := this.c.Users.Get(context.Background(), "")
	if err != nil {
		return "", err
	}
	return u.GetLogin(), err
}

func (this *githubClient) GetAuthorizedEmail() (string, error) {
	es, rs, err := this.c.Users.ListEmails(context.Background(), nil)
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
		if item.GetVerified() && item.GetPrimary() && item.GetVisibility() == "public" {
			return item.GetEmail(), nil
		}
	}

	return "", fmt.Errorf(errMsgNoPublicEmail)
}

func (this *githubClient) IsOrgExist(org string) (bool, error) {
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

func (this *githubClient) ListOrg() ([]string, error) {
	var r []string

	opt := github.ListOptions{PerPage: 99, Page: 1}
	for {
		ls, _, err := this.c.Organizations.List(context.Background(), "", &opt)
		if err != nil {
			return nil, err
		}

		if len(ls) == 0 {
			break
		}

		for _, v := range ls {
			r = append(r, v.GetLogin())
		}

		opt.Page += 1

	}

	return r, nil
}
