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

package oauth

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/oauth2"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	AuthApplyToLogin = "login"
	AuthApplyToSign  = "sign"
)

// key is the purpose that authorization applies to
var Auth = map[string]*codePlatformAuth{}

func Initialize(credentialFile string) error {
	cfg := authConfigs{}
	if err := util.LoadFromYaml(credentialFile, &cfg); err != nil {
		return err
	}

	f := func(purpose string, ac *authConfig) {
		cpa := &codePlatformAuth{
			webRedirectDir: ac.webRedirectDirConfig,
			clients:        map[string]AuthInterface{},
		}

		for _, item := range ac.Configs {
			cpa.clients[item.Platform] = &authClient{
				c: oauth2.NewOauth2Client(item.Oauth2Config),
			}
		}

		Auth[purpose] = cpa
	}

	f(AuthApplyToLogin, &cfg.Login)
	f(AuthApplyToSign, &cfg.Sign)
	return nil
}

type AuthInterface interface {
	GetAuthCodeURL(state string) string
	GetToken(code, scope string) (string, error)
	PasswordCredentialsToken(username, password string) (string, error)
}

type codePlatformAuth struct {
	webRedirectDir webRedirectDirConfig
	clients        map[string]AuthInterface
}

func (this *codePlatformAuth) GetAuthInstance(platform string) (AuthInterface, error) {
	if c, ok := this.clients[platform]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("Failed to get oauth instance: unknown platform: %s", platform)
}

func (this *codePlatformAuth) WebRedirectDir(success bool) string {
	if success {
		return this.webRedirectDir.WebRedirectDirOnSuccess
	}
	return this.webRedirectDir.WebRedirectDirOnFailure
}

type authClient struct {
	c oauth2.Oauth2Interface
}

func (this *authClient) GetAuthCodeURL(state string) string {
	return this.c.GetOauth2CodeURL(state)
}

func (this *authClient) GetToken(code, scope string) (string, error) {
	token, err := this.c.GetToken(code, scope)
	if err != nil {
		return "", fmt.Errorf("Get token failed: %s", err.Error())
	}

	return token.AccessToken, nil
}

func (this *authClient) PasswordCredentialsToken(username, password string) (string, error) {
	token, err := this.c.PasswordCredentialsToken(username, password)
	if err != nil {
		return "", fmt.Errorf("Get token failed: %s", err.Error())
	}

	return token.AccessToken, nil
}
