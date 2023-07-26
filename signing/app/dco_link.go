package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewDCOLinkService(
	repo repository.Link,
	dco claservice.CLAService,
	individual repository.IndividualSigning,
	emailCredential repository.EmailCredential,
) *dcoLinkService {
	return &dcoLinkService{
		repo:            repo,
		dco:             dco,
		individual:      individual,
		emailCredential: emailCredential,
	}
}

type DCOLinkService interface {
	Add(cmd *CmdToAddDCOLink) error
	Remove(linkId string) error
	List(cmd *CmdToListDCOLink) ([]repository.LinkSummary, error)
	Find(linkId string) (dto LinkDTO, err error)
	FindDCOs(string) ([]CLADetailDTO, error)
	FindDCO(cmd *domain.CLAIndex) (dto LinkCLADTO, err error)
}

type dcoLinkService struct {
	repo            repository.Link
	dco             claservice.CLAService
	individual      repository.IndividualSigning
	emailCredential repository.EmailCredential
}

func (s *dcoLinkService) Add(cmd *CmdToAddDCOLink) error {
	v, err := s.emailCredential.Find(cmd.Email)
	if err != nil {
		return err
	}

	link := cmd.toDCOLink()
	link.Email = domain.EmailInfo{
		Addr:     cmd.Email,
		Platform: v.Platform,
	}

	return s.dco.AddLink(&link)
}

func (s *dcoLinkService) Remove(linkId string) error {
	b, err := s.checkIfCanRemove(linkId)
	if err != nil {
		return err
	}
	if !b {
		return domain.NewDomainError(domain.ErrorCodeDCOLinkCanNotRemove)
	}

	v, err := s.repo.Find(linkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil
		}

		return err
	}

	return s.repo.Remove(&v)
}

func (s *dcoLinkService) checkIfCanRemove(linkId string) (bool, error) {
	v, err := s.individual.HasSignedLink(linkId)

	return !v, err
}

func (s *dcoLinkService) List(cmd *CmdToListDCOLink) ([]repository.LinkSummary, error) {
	return s.repo.FindAll(cmd)
}

func (s *dcoLinkService) FindDCOs(linkId string) ([]CLADetailDTO, error) {
	v, err := s.repo.Find(linkId)
	if err != nil {
		return nil, err
	}

	r := make([]CLADetailDTO, 0, len(v.CLAs))
	for i := range v.CLAs {
		item := &v.CLAs[i]

		r = append(r, CLADetailDTO{
			Id:       item.Id,
			Fileds:   item.Fields,
			Language: item.Language.Language(),
		})
	}

	return r, nil
}

func (s *dcoLinkService) FindDCO(cmd *domain.CLAIndex) (dto LinkCLADTO, err error) {
	v, err := s.repo.Find(cmd.LinkId)
	if err != nil {
		return
	}

	cla := v.FindCLA(cmd.CLAId)
	if cla == nil {
		err = domain.NewDomainError(domain.ErrorCodeDCONotExists)

		return
	}

	dto.Org = v.Org
	dto.Email = v.Email
	dto.CLA = CLADetailDTO{
		Id:        cla.Id,
		Fileds:    cla.Fields,
		Language:  cla.Language.Language(),
		LocalFile: s.dco.CLALocalFilePath(cmd),
	}

	return
}

func (s *dcoLinkService) Find(linkId string) (dto LinkDTO, err error) {
	v, err := s.repo.Find(linkId)
	if err != nil {
		return
	}

	dto.Org = v.Org
	dto.Email = v.Email

	return
}
