package email

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/zengchen1024/cla-server/models"
)

var emails = map[string]IEmail{}

type IEmail interface {
	GetOauth2CodeURL(state string) string
	GetAuthorizedEmail(code, scope string) (*models.OrgEmail, error)
	SendEmail(token oauth2.Token) error
	WebRedirectDir() string
	initialize(credentials, webRedirectDir string) error
}

func GetEmailClient(platform string) (IEmail, error) {
	e, ok := emails[platform]
	if !ok {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	return e, nil
}

func RegisterPlatform(platform, credentialFile, webRedirectDir string) error {
	e, err := GetEmailClient(platform)
	if err != nil {
		return err
	}
	return e.initialize(credentialFile, webRedirectDir)
}
