package models

type AccessToken struct {
	Id   string
	CSRF string
}

func NewAccessToken(payload []byte) (AccessToken, IModelError) {
	return accessTokenAdapterInstance.Add(payload)
}

func ValidateAndRefreshAccessToken(token AccessToken) (AccessToken, []byte, IModelError) {
	return accessTokenAdapterInstance.ValidateAndRefresh(token)
}
