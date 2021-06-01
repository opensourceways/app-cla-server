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

package obs

import (
	"fmt"

	appConf "github.com/opensourceways/app-cla-server/config"
)

type OBS interface {
	Initialize(string, string) error
	WriteObject(path string, data []byte) error
	ReadObject(path, localPath string) OBSError
	HasObject(string) (bool, error)
	ListObject(pathPrefix string) ([]string, error)
}

var instances = map[string]OBS{}

func Register(plugin string, i OBS) {
	instances[plugin] = i
}

func Initialize(info appConf.OBS) (OBS, error) {
	i, ok := instances[info.Name]
	if !ok {
		return nil, fmt.Errorf("no such obs instance of %s", info.Name)
	}

	return i, i.Initialize(info.CredentialFile, info.Bucket)
}

type OBSError interface {
	Error() string
	IsObjectNotFound() bool
}
