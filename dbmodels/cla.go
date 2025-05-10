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

const (
	ApplyToCorporation = "corporation"
	ApplyToIndividual  = "individual"
)

type CLA struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Text     string  `json:"text"`
	Language string  `json:"language"`
	Fields   []Field `json:"fields"`
}

type Field struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type CLAListOptions struct {
	Submitter string `json:"submitter"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	ApplyTo   string `json:"apply_to"`
}

type CLAData struct {
	URL      string  `json:"url"`
	Language string  `json:"language"`
	Fields   []Field `json:"fields"`
}

type CLADetail struct {
	CLAData
	CLAHash string `json:"cla_hash"`
	Text    string `json:"text"`
}

type CLACreateOption struct {
	OrgSignature     *[]byte `json:"org_signature"`
	OrgSignatureHash string

	CLADetail
}

type CLAInfo struct {
	CLALang          string
	CLAHash          string
	OrgSignatureHash string
	Fields           []Field
}
