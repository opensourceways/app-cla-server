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

package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningBasicInfo struct {
	CLALanguage     string `json:"cla_language"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	AdminAdded bool `json:"admin_added"`
}

type CorpSigningCreateOpt struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}
