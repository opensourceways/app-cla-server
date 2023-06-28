package app

import "github.com/opensourceways/app-cla-server/signing/domain/repository"

func NewCorpEmailDomainService(
	repo repository.CorpSigning,
) CorpEmailDomainService {
	return &corpEmailDomainService{
		repo: repo,
	}
}

type CorpEmailDomainService interface {
	Add(cmd *CmdToAddEmailDomain) error
	List(string) ([]string, error)
}

type corpEmailDomainService struct {
	repo repository.CorpSigning
}

func (s *corpEmailDomainService) Add(cmd *CmdToAddEmailDomain) error {
	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return err
	}

	if err := cs.AddEmailDomain(cmd.EmailAddr); err != nil {
		return err
	}

	return s.repo.AddEmailDomain(&cs, cmd.EmailAddr.Domain())
}

func (s *corpEmailDomainService) List(csId string) ([]string, error) {
	return s.repo.FindEmailDomains(csId)
}
