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

import "github.com/opensourceways/app-cla-server/oauth2"

type authConfigs struct {
	Login authConfig `json:"login" required:"true"`
	Sign  authConfig `json:"sign" required:"true"`
}

type authConfig struct {
	webRedirectDirConfig

	Configs []platformConfig `json:"platforms" required:"true"`
}

type platformConfig struct {
	Platform string `json:"platform" required:"true"`

	oauth2.Oauth2Config
}

type webRedirectDirConfig struct {
	WebRedirectDirOnSuccess string `json:"web_redirect_dir_on_success" required:"true"`
	WebRedirectDirOnFailure string `json:"web_redirect_dir_on_failure" required:"true"`
}
