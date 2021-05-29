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

package huaweicloud

import (
	"github.com/opensourceways/app-cla-server/util"
)

type config struct {
	AccessKey           string `json:"access_key" required:"true"`
	SecretKey           string `json:"secret_key" required:"true"`
	Endpoint            string `json:"endpoint" required:"true"`
	ObjectEncryptionKey string `json:"object_encryption_key"`
}

func loadConfig(path string) (*config, error) {
	v := &config{}
	if err := util.LoadFromYaml(path, v); err != nil {
		return nil, err
	}

	return v, nil
}
