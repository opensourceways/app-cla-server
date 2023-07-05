package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type AccessToken interface {
	Add(*domain.AccessTokenDO) (string, error)
	Find(string) (domain.AccessTokenDO, error)
	Delete(string) error
}
