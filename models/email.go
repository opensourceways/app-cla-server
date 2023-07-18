package models

type EmailAuthorizationReq struct {
	Email     string `json:"email"`
	Authorize []byte `json:"authorize"`
}

func (req *EmailAuthorizationReq) Clear() {
	for i := range req.Authorize {
		req.Authorize[i] = 0
	}
}

type EmailAuthorization struct {
	Code string `json:"code"`
	EmailAuthorizationReq
}
