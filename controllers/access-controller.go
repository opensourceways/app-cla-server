package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/util"
)

const (
	PermissionOwnerOfOrg       = "owner of org"
	PermissionIndividualSigner = "individual signer"
	PermissionCorpAdmin        = "corporation administrator"
	PermissionEmployeeManager  = "employee manager"
)

type accessController struct {
	Expiry     int64       `json:"expiry"`
	Permission string      `json:"permission"`
	Payload    interface{} `json:"payload"`
}

func (this *accessController) newToken(secret string) (string, error) {
	body, err := golangsdk.BuildRequestBody(this, "")
	if err != nil {
		return "", fmt.Errorf("Failed to create token: build body failed: %s", err.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims(body)

	return token.SignedString([]byte(secret))
}

func (this *accessController) refreshToken(expiry int64, secret string) (string, error) {
	this.Expiry = util.Expiry(expiry)
	return this.newToken(secret)
}

func (this *accessController) parseToken(token, secret string) error {
	t, err := jwt.Parse(token, func(t1 *jwt.Token) (interface{}, error) {
		if _, ok := t1.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !t.Valid {
		return fmt.Errorf("Not a valid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("Not valid claims")
	}

	d, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	return json.Unmarshal(d, this)
}

func (this *accessController) isTokenExpired() bool {
	return this.Expiry < util.Now()
}

func (this *accessController) verify(permission []string) error {
	for _, item := range permission {
		if this.Permission == item {
			return nil
		}
	}

	return fmt.Errorf("Not allowed permission")
}
