package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type IndividualSigning interface {
	Add(*domain.IndividualSigning) error
	AddForMigrate(*domain.IndividualSigning) error
	FindSignedCLA(linkId string, email dp.EmailAddr) (claId string, language dp.Language, err error)
	Find(linkId string, email dp.EmailAddr) (domain.IndividualSigning, error)
	FindAll(LinkId string) ([]domain.IndividualSigning, error)
	FindAllWithPagination(linkId string, offset, limit int) ([]domain.IndividualSigning, error)
	CountByLinkId(linkId string) (int64, error)

	HasSignedLink(linkId string) (bool, error)
	HasSignedCLA(*domain.CLAIndex) (bool, error)
	SaveNewCLA(is *domain.IndividualSigning) error
	UpdateCLANotify(is *domain.IndividualSigning) error
}
