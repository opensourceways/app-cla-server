package controllers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/huaweicloud/golangsdk"

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
		return "", fmt.Errorf("Failed to create token: build body failed: %s", err.Error())
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

func (this *accessController) verify(permission []string, addr string) error {
	bingo := false
	for _, item := range permission {
		if this.Permission == item {
			bingo = true
			break
		}
	}
	if !bingo {
		return fmt.Errorf("Not allowed permission")
	}

	if this.RemoteAddr != addr {
		return fmt.Errorf("Unmatched remote address")
	}
	return nil
}

func (this *accessController) symmetricEncryptionKey() []byte {
	return []byte(config.AppConfig.SymmetricEncryptionKey)
}

func (this *accessController) encryptToken(token string) (string, error) {
	t, err := util.Encrypt([]byte(token), this.symmetricEncryptionKey())
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

	s, err := util.Decrypt(dst, this.symmetricEncryptionKey())
	if err != nil {
		return "", err
	}

	return string(s), nil
}
