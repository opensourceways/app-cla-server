package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/huaweicloud/golangsdk"
)

const (
	PermissionOwnerOfOrg       = "owner of org"
	PermissionIndividualSigner = "individual signer"
	PermissionCorporAdmin      = "corporation administrator"
	PermissionEmployeeManager  = "employee manager"
)

type accessControllerInterface interface {
	NewToken(int64) (string, error)
	ParseToken(token, secret string) error
	Verify(permission []string) error
	GetUser() string
}

type accessController struct {
	Expiry     int64  `json:"expiry"`
	User       string `json:"user"`
	Permission string `json:"permission"`
	secret     string `json:"-"`
}

type codePlatformAuth struct {
	accessController

	PlatformToken string `json:"platform_token"`
}

func (this *accessController) NewToken(expiry int64) (string, error) {
	this.Expiry = time.Now().Add(time.Second * time.Duration(expiry)).Unix()

	body, err := golangsdk.BuildRequestBody(this, "")
	if err != nil {
		return "", fmt.Errorf("Failed to create token: build body failed: %s", err.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims(body)

	return token.SignedString([]byte(this.secret))
}

func (this *accessController) ParseToken(token, secret string) error {
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

	if err = json.Unmarshal(d, this); err != nil {
		return err
	}

	this.secret = secret
	return nil
}

func (this *accessController) GetUser() string {
	return this.User
}

func (this *accessController) Verify(permission []string) error {
	if this.Expiry < time.Now().Unix() {
		return fmt.Errorf("token is expired")
	}

	for _, item := range permission {
		if this.Permission == item {
			return nil
		}
	}

	return fmt.Errorf("Not allowed permission")
}
