package models

type EmailAuthorizationReq struct {
	Email     string `json:"email"`
	Authorize string `json:"authorize"`
}

type EmailAuthorization struct {
	Code string `json:"code"`
	EmailAuthorizationReq
}
