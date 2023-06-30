package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type IndividualSigning interface {
	Add(*domain.IndividualSigning) error
	Count(linkId string, email dp.EmailAddr) (int, error)
}
