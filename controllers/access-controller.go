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
	PermissionCorporAdmin      = "corporation administrator"
	PermissionEmployeeManager  = "employee manager"
)

type accessController struct {
	Expiry     int64       `json:"expiry"`
	Permission string      `json:"permission"`
	Payload    interface{} `json:"payload"`
}

func (this *accessController) NewToken(secret string) (string, error) {
	body, err := golangsdk.BuildRequestBody(this, "")
	if err != nil {
		return "", fmt.Errorf("Failed to create token: build body failed: %s", err.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims(body)

	return token.SignedString([]byte(secret))
}

func (this *accessController) RefreshToken(expiry int64, secret string) (string, error) {
	this.Expiry = util.Expiry(expiry)
	return this.NewToken(secret)
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

	if this.Expiry < util.Now() {
		return fmt.Errorf("token is expired")
	}

	return nil
}

func (this *accessController) Verify(permission []string) error {
	for _, item := range permission {
		if this.Permission == item {
			return nil
		}
	}

	return fmt.Errorf("Not allowed permission")
}

type acForCodePlatformPayload struct {
	User          string          `json:"user"`
	Email         string          `json:"email"`
	Platform      string          `json:"platform"`
	PlatformToken string          `json:"platform_token"`
	Orgs          map[string]bool `json:"orgs"`
}

func (this *acForCodePlatformPayload) hasOrg(org string) bool {
	if this.Orgs == nil {
		return false
	}

	_, ok := this.Orgs[org]
	return ok
}

func (this *acForCodePlatformPayload) addOrg(org string) {
	if this.Orgs == nil {
		this.Orgs = map[string]bool{}
	}

	this.Orgs[org] = true
}

type acForCorpManagerPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	OrgCLAID string `json:"org_cla_id"`
}

func (this *acForCorpManagerPayload) hasEmployee(email string) bool {
	return util.EmailSuffix(this.Email) == util.EmailSuffix(email)
}
