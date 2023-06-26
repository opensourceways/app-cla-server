package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type User interface {
	Add(*domain.User) (string, error)
	Remove(string) error
	SavePassword(*domain.User) error
	Find(string) (domain.User, error)
}
