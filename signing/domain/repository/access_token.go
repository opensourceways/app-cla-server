package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type AccessToken interface {
	Add(*domain.AccessToken) (string, error)
	Find(string) (domain.AccessToken, error)
	Delete(string) error
}
