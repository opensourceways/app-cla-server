package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type IndividualSigning interface {
	Add(*domain.IndividualSigning) error
	FindSignedCLA(linkId string, email dp.EmailAddr) (string, error)
	Find(linkId string, email dp.EmailAddr) (domain.IndividualSigning, error)

	HasSignedLink(linkId string) (bool, error)
	HasSignedCLA(*domain.CLAIndex) (bool, error)
	SaveNewCLA(is *domain.IndividualSigning) error
}
