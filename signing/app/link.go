package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewLinkService(
	repo repository.Link,
	cla claservice.CLAService,
	cs repository.CorpSigning,
	individual repository.IndividualSigning,
	emailCredential repository.EmailCredential,
) *linkService {
	return &linkService{
		repo:            repo,
		cla:             cla,
		cs:              cs,
		individual:      individual,
		emailCredential: emailCredential,
	}
}

type LinkService interface {
	Add(cmd *CmdToAddLink) error
	Remove(userId, linkId string) error
	List(userId string) ([]repository.LinkSummary, error)
	Find(linkId string) (dto LinkDTO, err error)
	FindCLAs(cmd *CmdToFindCLAs) ([]CLADetailDTO, error)
	FindLinkCLA(cmd *domain.CLAIndex) (dto LinkCLADTO, err error)
}

type linkService struct {
	repo            repository.Link
	cla             claservice.CLAService
	cs              repository.CorpSigning
	individual      repository.IndividualSigning
	emailCredential repository.EmailCredential
}

func (s *linkService) Add(cmd *CmdToAddLink) error {
	v, err := s.emailCredential.Find(cmd.Email)
	if err != nil {
		return err
	}

	link := cmd.toLink()
	link.Email = domain.EmailInfo{
		Addr:     cmd.Email,
		Platform: v.Platform,
	}

	return s.cla.AddLink(&link)
}

func (s *linkService) Remove(userId, linkId string) error {
	b, err := s.checkIfCanRemove(linkId)
	if err != nil {
		return err
	}
	if !b {
		return domain.NewDomainError(domain.ErrorCodeLinkCanNotRemove)
	}

	v, err := checkIfCommunityManager(userId, linkId, s.repo)
	if err != nil {
		if domain.IsErrorOf(err, domain.ErrorCodeLinkNotExists) {
			return nil
		}

		return err
	}

	return s.repo.Remove(v)
}

func (s *linkService) checkIfCanRemove(linkId string) (bool, error) {
	v, err := s.individual.HasSignedLink(linkId)
	if err != nil {
		return v, err
	}
	if v {
		return false, nil
	}

	v, err = s.cs.HasSignedLink(linkId)

	return !v, err
}

func (s *linkService) List(userId string) ([]repository.LinkSummary, error) {
	return s.repo.FindAll(userId)
}

func (s *linkService) FindCLAs(cmd *CmdToFindCLAs) ([]CLADetailDTO, error) {
	v, err := s.repo.Find(cmd.LinkId)
	if err != nil {
		return nil, err
	}

	t := cmd.Type.CLAType()

	r := make([]CLADetailDTO, 0, len(v.CLAs))
	for i := range v.CLAs {
		item := &v.CLAs[i]

		if item.Type.CLAType() == t {
			r = append(r, CLADetailDTO{
				Id:       item.Id,
				Fileds:   item.Fields,
				Language: item.Language.Language(),
			})
		}
	}

	return r, nil
}

func (s *linkService) FindLinkCLA(cmd *domain.CLAIndex) (dto LinkCLADTO, err error) {
	v, err := s.repo.Find(cmd.LinkId)
	if err != nil {
		return
	}

	cla := v.FindCLA(cmd.CLAId)
	if cla == nil {
		err = domain.NewDomainError(domain.ErrorCodeCLANotExists)

		return
	}

	dto.Org = v.Org
	dto.Email = v.Email
	dto.CLA = CLADetailDTO{
		Id:        cla.Id,
		Fileds:    cla.Fields,
		Language:  cla.Language.Language(),
		LocalFile: s.cla.CLALocalFilePath(cmd),
	}

	return
}

func (s *linkService) Find(linkId string) (dto LinkDTO, err error) {
	v, err := s.repo.Find(linkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeLinkNotExists)
		}

		return
	}

	dto.Org = v.Org
	dto.Email = v.Email

	return
}
