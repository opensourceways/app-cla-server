package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type Login interface {
	Add(*domain.Login) error
	Find(string) (domain.Login, error)
	Delete(string) error
}
