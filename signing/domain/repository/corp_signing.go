package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type CorpSigning interface {
	Add(*domain.CorpSigning) error
	AddEmployee(*domain.CorpSigning) error
	AddAdmin(*domain.CorpSigning) error

	// count the corp by the email domain
	Count(domain string) (int, error)
	Find(string) (domain.CorpSigning, error)
}
