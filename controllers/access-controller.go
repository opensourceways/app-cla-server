package controllers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/huaweicloud/golangsdk"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	PermissionOwnerOfOrg       = "owner of org"
	PermissionIndividualSigner = "individual signer"
	PermissionCorpAdmin        = "corporation administrator"
	PermissionEmployeeManager  = "employee manager"
)

type accessController struct {
	RemoteAddr string      `json:"remote_addr"`
	Expiry     int64       `json:"expiry"`
	Permission string      `json:"permission"`
	Payload    interface{} `json:"payload"`
}

func (this *accessController) newToken(secret string) (string, error) {
	body, err := golangsdk.BuildRequestBody(this, "")
	if err != nil {
		return "", fmt.Errorf("failed to create token: build body failed: %s", err.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims(body)

	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return this.encryptToken(s)
}

func (this *accessController) refreshToken(expiry int64, secret string) (string, error) {
	this.Expiry = util.Expiry(expiry)
	return this.newToken(secret)
}

func (this *accessController) parseToken(token, secret string) error {
	token1, err := this.decryptToken(token)
	if err != nil {
		return err
	}

	t, err := jwt.Parse(token1, func(t1 *jwt.Token) (interface{}, error) {
		if _, ok := t1.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !t.Valid {
		return fmt.Errorf("not a valid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("not valid claims")
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

func (this *accessController) verify(permission []string, addr string) error {
	if !sets.NewString(permission...).Has(this.Permission) {
		return fmt.Errorf("not allowed permission")
	}

	if this.RemoteAddr != addr {
		return fmt.Errorf("unmatched remote address")
	}

	return nil
}

func (this *accessController) newEncryption() util.SymmetricEncryption {
	e, _ := util.NewSymmetricEncryption(config.AppConfig.SymmetricEncryptionKey, "")
	return e
}

func (this *accessController) encryptToken(token string) (string, error) {
	t, err := this.newEncryption().Encrypt([]byte(token))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(t), nil
}

func (this *accessController) decryptToken(token string) (string, error) {
	dst, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}

	s, err := this.newEncryption().Decrypt(dst)
	if err != nil {
		return "", err
	}

	return string(s), nil
}
