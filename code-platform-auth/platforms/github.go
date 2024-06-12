/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func (gc *githubClient) HasRepo(org, repo string) (bool, error) {
	_, r, err := gc.c.Repositories.Get(context.Background(), org, repo)
	if err == nil {
		return true, nil
	}

	if r.StatusCode == 404 {
		return false, nil
	}

	return false, err
}
