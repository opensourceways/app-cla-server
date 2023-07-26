package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewDCOService(
	repo repository.Link,
	dco claservice.CLAService,
	individual repository.IndividualSigning,
) *dcoService {
	return &dcoService{
		repo:       repo,
		dco:        dco,
		individual: individual,
	}
}

type DCOService interface {
	Add(linkId string, cmd *CmdToAddDCO) error
	Remove(cmd domain.CLAIndex) error
	DCOLocalFilePath(domain.CLAIndex) string
	List(linkId string) ([]CLADTO, error)
}

type dcoService struct {
	repo       repository.Link
	dco        claservice.CLAService
	individual repository.IndividualSigning
}

func (s *dcoService) Add(linkId string, cmd *CmdToAddDCO) error {
	link, err := s.repo.Find(linkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeDCOLinkNotExists)
		}

		return err
	}

	v := cmd.toDCO()

	return s.dco.Add(&link, &v)
}

func (s *dcoService) Remove(cmd domain.CLAIndex) error {
	link, err := s.repo.Find(cmd.LinkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeDCOLinkNotExists)
		}

		return err
	}

	cla := link.FindCLA(cmd.CLAId)
	if cla == nil {
		return domain.NewDomainError(domain.ErrorCodeDCONotExists)
	}

	if err := s.checkIfCanRemove(&cmd); err != nil {
		return err
	}

	return s.repo.RemoveCLA(&link, cla)
}

func (s *dcoService) checkIfCanRemove(cmd *domain.CLAIndex) error {
	v, err := s.individual.HasSignedCLA(cmd)
	if err != nil {
		return err
	}
	if v {
		return domain.NewDomainError(domain.ErrorCodeDCOCanNotRemove)
	}

	return nil
}

func (s *dcoService) List(linkId string) (dcos []CLADTO, err error) {
	v, err := s.repo.Find(linkId)
	if err != nil {
		return
	}

	dcos = make([]CLADTO, 0, len(v.CLAs))

	for i := range v.CLAs {
		item := &v.CLAs[i]

		dcos = append(dcos, CLADTO{
			Id:       item.Id,
			URL:      item.URL.URL(),
			Language: item.Language.Language(),
		})
	}

	return
}

func (s *dcoService) DCOLocalFilePath(index domain.CLAIndex) string {
	return s.dco.CLALocalFilePath(&index)
}
