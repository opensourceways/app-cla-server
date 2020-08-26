package email

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/zengchen1024/cla-server/models"
)

type IEmail interface {
	GetOauth2CodeURL(state string) string
	GetAuthorizedEmail(code, scope string) (*models.OrgEmail, error)
	SendEmail(token oauth2.Token) error
}

func GetEmailClient(platform string) (IEmail, error) {
	if platform != "gmail" {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	return gmailCli, nil
}
