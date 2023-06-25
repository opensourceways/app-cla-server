package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type CorpSigning interface {
	Add(*domain.CorpSigning) error
	AddEmployee(*domain.CorpSigning) error

	Find(string) (domain.CorpSigning, error)
}
