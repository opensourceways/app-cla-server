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
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerCreateOption struct {
	ID       string
	Name     string
	Role     string
	Email    string
	Password string
}

type CorporationManagerCheckInfo struct {
	ID          string
	Email       string
	EmailSuffix string
}

type CorporationManagerResetPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CorporationManagerCheckResult struct {
	Role             string
	Name             string
	Email            string
	Password         string
	InitialPWChanged bool

	OrgInfo
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
