package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCLAService(
	repo repository.Link,
	cla claservice.CLAService,
) *claService {
	return &claService{
		repo: repo,
		cla:  cla,
	}
}

type CLAService interface {
	Add(linkId string, cmd *CmdToAddCLA) error
	Remove(cmd domain.CLAIndex) error
	CLALocalFilePath(domain.CLAIndex) string
	List(linkId string) ([]CLADTO, []CLADTO, error)
}

type claService struct {
	repo repository.Link
	cla  claservice.CLAService
}

func (s *claService) Add(linkId string, cmd *CmdToAddCLA) error {
	link, err := s.repo.Find(linkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeLinkNotExists)
		}

		return err
	}

	cla := cmd.toCLA()

	return s.cla.Add(&link, &cla)
}

func (s *claService) Remove(cmd domain.CLAIndex) error {
	link, err := s.repo.Find(cmd.LinkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeLinkNotExists)
		}

		return err
	}

	cla := link.FindCLA(cmd.CLAId)
	if cla == nil {
		return domain.NewDomainError(domain.ErrorCodeCLANotExists)
	}

	// TODO can't delete if it is being used

	return s.repo.RemoveCLA(&link, cla)
}

func (s *claService) List(linkId string) (individuals []CLADTO, corps []CLADTO, err error) {
	v, err := s.repo.Find(linkId)
	if err != nil {
		return
	}

	corps = make([]CLADTO, 0, len(v.CLAs))
	individuals = make([]CLADTO, 0, len(v.CLAs))

	for i := range v.CLAs {
		item := &v.CLAs[i]

		dto := CLADTO{
			Id:       item.Id,
			URL:      item.URL.URL(),
			Type:     item.Type.CLAType(),
			Language: item.Language.Language(),
		}

		if dp.IsCLATypeIndividual(item.Type) {
			individuals = append(individuals, dto)
		} else {
			corps = append(corps, dto)
		}
	}

	return
}

func (s *claService) CLALocalFilePath(index domain.CLAIndex) string {
	return s.cla.CLALocalFilePath(&index)
}
